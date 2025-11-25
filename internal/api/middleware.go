package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/internal/ratelimit"
	"github.com/infiniv/rsearch/internal/validation"
)

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(cfg.Features.RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}
			w.Header().Set(cfg.Features.RequestIDHeader, requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *observability.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				requestID := w.Header().Get("X-Request-ID")

				logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"method":     r.Method,
					"path":       r.URL.Path,
					"status":     ww.Status(),
					"bytes":      ww.BytesWritten(),
					"duration":   duration.Milliseconds(),
					"remote":     r.RemoteAddr,
				}).Infof("%s %s %d %dms", r.Method, r.URL.Path, ww.Status(), duration.Milliseconds())
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *observability.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := w.Header().Get("X-Request-ID")
					logger.WithFields(map[string]interface{}{
						"request_id": requestID,
						"panic":      err,
					}).Error("Panic recovered")

					RespondInternalError(w, "An unexpected error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.CORS.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			if origin != "" {
				// Check if origin is allowed
				allowed := false
				for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}

				if !allowed {
					next.ServeHTTP(w, r)
					return
				}

				// Set other CORS headers
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				// Handle preflight
				if r.Method == "OPTIONS" {
					methods := ""
					for i, method := range cfg.CORS.AllowedMethods {
						if i > 0 {
							methods += ", "
						}
						methods += method
					}
					w.Header().Set("Access-Control-Allow-Methods", methods)
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
					w.Header().Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware records metrics for requests
func MetricsMiddleware(metrics *observability.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			metrics.IncActiveRequests()
			defer metrics.DecActiveRequests()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			metrics.RecordRequest(r.URL.Path, ww.Status(), duration)
		})
	}
}

// ValidationMiddleware validates request body size and enforces security constraints
func ValidationMiddleware(validator *validation.Validator, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST requests (translate, schema registration)
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Validate request body size
			if r.ContentLength > cfg.Limits.MaxRequestBodySize {
				RespondError(w, http.StatusBadRequest, "REQUEST_TOO_LARGE",
					"Request body exceeds maximum allowed size")
				return
			}

			// Use MaxBytesReader to enforce body size limit
			r.Body = http.MaxBytesReader(w, r.Body, cfg.Limits.MaxRequestBodySize)

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implements per-IP rate limiting
func RateLimitMiddleware(limiter *ratelimit.RateLimiter, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if disabled
			if !cfg.Limits.RateLimit.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract client IP
			clientIP := extractClientIP(r)

			// Check rate limit
			if !limiter.Allow(clientIP) {
				// Calculate retry after based on requests per minute
				retryAfterSeconds := 60 / cfg.Limits.RateLimit.RequestsPerMinute
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
				RespondError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					"Rate limit exceeded. Please try again later.")
				return
			}

			// Add rate limit remaining header (informational)
			w.Header().Set("X-RateLimit-Remaining", "available")

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP extracts the client IP from the request
// Priority: X-Forwarded-For > X-Real-IP > RemoteAddr
func extractClientIP(r *http.Request) string {
	// Try X-Forwarded-For first (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP from the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port if present)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, RemoteAddr might not have a port
		return r.RemoteAddr
	}

	return ip
}

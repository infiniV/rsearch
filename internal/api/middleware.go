package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
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

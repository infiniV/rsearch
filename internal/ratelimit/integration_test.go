package ratelimit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infiniv/rsearch/internal/config"
)

// TestRateLimitIntegration tests rate limiting in a realistic scenario
func TestRateLimitIntegration(t *testing.T) {
	// Create a test handler that counts successful requests
	successCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Request %d", successCount)
	})

	// Create rate limiter: 60 req/min with burst of 5
	limiter := NewRateLimiter(60, 5)
	defer limiter.Stop()

	// Create config with rate limiting enabled
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             5,
			},
		},
	}

	// Wrap handler with rate limit middleware
	rateLimitedHandler := applyRateLimitMiddleware(handler, limiter, cfg)

	// Test scenario: burst of 10 requests from same IP
	ip := "192.168.1.100"
	allowedCount := 0
	blockedCount := 0

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = fmt.Sprintf("%s:12345", ip)
		rr := httptest.NewRecorder()

		rateLimitedHandler.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			allowedCount++
		} else if rr.Code == http.StatusTooManyRequests {
			blockedCount++

			// Verify Retry-After header is present
			if rr.Header().Get("Retry-After") == "" {
				t.Errorf("request %d: missing Retry-After header on 429 response", i+1)
			}
		}
	}

	if allowedCount != 5 {
		t.Errorf("expected 5 allowed requests, got %d", allowedCount)
	}

	if blockedCount != 5 {
		t.Errorf("expected 5 blocked requests, got %d", blockedCount)
	}

	// Verify different IP has independent quota
	ip2 := "192.168.1.101"
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:12345", ip2)
	rr := httptest.NewRecorder()

	rateLimitedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("different IP should have independent quota, got status %d", rr.Code)
	}
}

// applyRateLimitMiddleware is a helper to apply rate limiting middleware
func applyRateLimitMiddleware(handler http.Handler, limiter *RateLimiter, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Limits.RateLimit.Enabled {
			handler.ServeHTTP(w, r)
			return
		}

		// Extract client IP (simplified for testing)
		ip := r.RemoteAddr
		if idx := len(ip) - 1; idx >= 0 {
			for i := len(ip) - 1; i >= 0; i-- {
				if ip[i] == ':' {
					ip = ip[:i]
					break
				}
			}
		}

		if !limiter.Allow(ip) {
			retryAfterSeconds := 60 / cfg.Limits.RateLimit.RequestsPerMinute
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", "available")
		handler.ServeHTTP(w, r)
	})
}

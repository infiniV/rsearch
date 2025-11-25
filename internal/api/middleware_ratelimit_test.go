package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/ratelimit"
)

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled: false,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(10, 5)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests - all should succeed since rate limiting is disabled
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             5,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests within burst limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}

		// Check for rate limit remaining header
		remaining := rr.Header().Get("X-RateLimit-Remaining")
		if remaining == "" {
			t.Errorf("request %d: missing X-RateLimit-Remaining header", i+1)
		}
	}
}

func TestRateLimitMiddleware_BlocksExceededLimit(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             3,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests exceeding burst limit
	successCount := 0
	blockedCount := 0

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			successCount++
		} else if rr.Code == http.StatusTooManyRequests {
			blockedCount++

			// Check for Retry-After header
			retryAfter := rr.Header().Get("Retry-After")
			if retryAfter == "" {
				t.Errorf("request %d: missing Retry-After header on 429 response", i+1)
			}
		} else {
			t.Errorf("request %d: unexpected status code %d", i+1, rr.Code)
		}
	}

	if successCount != 3 {
		t.Errorf("expected 3 successful requests, got %d", successCount)
	}

	if blockedCount != 2 {
		t.Errorf("expected 2 blocked requests, got %d", blockedCount)
	}
}

func TestRateLimitMiddleware_ExtractsIPFromXForwardedFor(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             2,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP exhausts its limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if i < 2 && rr.Code != http.StatusOK {
			t.Errorf("IP1 request %d: expected status 200, got %d", i+1, rr.Code)
		}
		if i >= 2 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("IP1 request %d: expected status 429, got %d", i+1, rr.Code)
		}
	}

	// Different IP should still have its full quota
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.3")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("IP2 request: expected status 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_ExtractsIPFromXRealIP(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             2,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use X-Real-IP header
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "10.0.0.5")
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if i < 2 && rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}
		if i >= 2 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("request %d: expected status 429, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimitMiddleware_UsesRemoteAddrFallback(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             2,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No X-Forwarded-For or X-Real-IP, should use RemoteAddr
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if i < 2 && rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, rr.Code)
		}
		if i >= 2 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("request %d: expected status 429, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimitMiddleware_RetryAfterHeader(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60, // 1 token per second
				Burst:             1,
			},
		},
	}

	limiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer limiter.Stop()

	middleware := RateLimitMiddleware(limiter, cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("first request should succeed, got status %d", rr1.Code)
	}

	// Next request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got status %d", rr2.Code)
	}

	retryAfter := rr2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Fatal("Retry-After header should be present")
	}

	// Retry-After should be a positive integer (seconds)
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		t.Fatalf("Retry-After should be an integer, got: %s", retryAfter)
	}

	if seconds < 1 || seconds > 60 {
		t.Errorf("Retry-After should be between 1 and 60 seconds, got: %d", seconds)
	}
}

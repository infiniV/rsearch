package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name               string
		requestsPerMinute  int
		burst              int
		ip                 string
		requests           int
		expectAllowedCount int
	}{
		{
			name:               "allows requests within burst",
			requestsPerMinute:  60,
			burst:              5,
			ip:                 "192.168.1.1",
			requests:           5,
			expectAllowedCount: 5,
		},
		{
			name:               "blocks requests exceeding burst",
			requestsPerMinute:  60,
			burst:              3,
			ip:                 "192.168.1.2",
			requests:           5,
			expectAllowedCount: 3,
		},
		{
			name:               "allows single request with zero burst",
			requestsPerMinute:  60,
			burst:              0,
			ip:                 "192.168.1.3",
			requests:           1,
			expectAllowedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.requestsPerMinute, tt.burst)
			defer limiter.Stop()

			allowed := 0
			for i := 0; i < tt.requests; i++ {
				if limiter.Allow(tt.ip) {
					allowed++
				}
			}

			if allowed != tt.expectAllowedCount {
				t.Errorf("expected %d requests allowed, got %d", tt.expectAllowedCount, allowed)
			}
		})
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	requestsPerMinute := 60
	burst := 2
	limiter := NewRateLimiter(requestsPerMinute, burst)
	defer limiter.Stop()

	ip := "192.168.1.10"

	// Consume all tokens in burst
	if !limiter.Allow(ip) {
		t.Fatal("first request should be allowed")
	}
	if !limiter.Allow(ip) {
		t.Fatal("second request should be allowed")
	}

	// Third request should be blocked (burst exhausted)
	if limiter.Allow(ip) {
		t.Fatal("third request should be blocked")
	}

	// Wait for token refill (60 req/min = 1 req/sec)
	// Wait slightly more than 1 second to ensure token is refilled
	time.Sleep(1100 * time.Millisecond)

	// Should now allow one more request
	if !limiter.Allow(ip) {
		t.Error("request should be allowed after token refill")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	requestsPerMinute := 100
	burst := 10
	limiter := NewRateLimiter(requestsPerMinute, burst)
	defer limiter.Stop()

	// Test concurrent access from multiple IPs
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	var wg sync.WaitGroup

	for _, ip := range ips {
		ip := ip // capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				limiter.Allow(ip)
			}
		}()
	}

	// Should not panic or deadlock
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test passed
	case <-time.After(5 * time.Second):
		t.Fatal("test timed out - possible deadlock")
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	requestsPerMinute := 60
	burst := 3
	limiter := NewRateLimiter(requestsPerMinute, burst)
	defer limiter.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// IP1 exhausts its burst
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip1) {
			t.Fatalf("IP1 request %d should be allowed", i+1)
		}
	}

	// IP1 should be blocked
	if limiter.Allow(ip1) {
		t.Error("IP1 should be rate limited")
	}

	// IP2 should still have its full burst available
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip2) {
			t.Fatalf("IP2 request %d should be allowed", i+1)
		}
	}

	// IP2 should now be blocked
	if limiter.Allow(ip2) {
		t.Error("IP2 should be rate limited")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Use a shorter cleanup interval for testing
	requestsPerMinute := 60
	burst := 5
	limiter := NewRateLimiterWithCleanup(requestsPerMinute, burst, 100*time.Millisecond, 200*time.Millisecond)
	defer limiter.Stop()

	ip := "192.168.1.100"

	// Make a request to create bucket
	if !limiter.Allow(ip) {
		t.Fatal("first request should be allowed")
	}

	// Verify bucket exists
	_, exists := limiter.buckets.Load(ip)
	if !exists {
		t.Fatal("bucket should exist after request")
	}

	// Wait for cleanup cycle (stale threshold + cleanup interval)
	time.Sleep(400 * time.Millisecond)

	// Bucket should be cleaned up
	_, exists = limiter.buckets.Load(ip)
	if exists {
		t.Error("stale bucket should have been cleaned up")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	limiter := NewRateLimiter(60, 5)

	// Stop should not panic
	limiter.Stop()

	// Calling Stop again should not panic
	limiter.Stop()
}

func TestRateLimiter_BurstHandling(t *testing.T) {
	requestsPerMinute := 60 // 1 token per second
	burst := 5
	limiter := NewRateLimiter(requestsPerMinute, burst)
	defer limiter.Stop()

	ip := "192.168.1.20"

	// Burst should allow 5 immediate requests
	for i := 1; i <= burst; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("burst request %d/%d should be allowed", i, burst)
		}
	}

	// 6th request should be blocked
	if limiter.Allow(ip) {
		t.Error("request exceeding burst should be blocked")
	}

	// Wait for 2 seconds to get 2 tokens back
	time.Sleep(2100 * time.Millisecond)

	// Should allow 2 more requests
	if !limiter.Allow(ip) {
		t.Error("request after refill should be allowed")
	}
	if !limiter.Allow(ip) {
		t.Error("second request after refill should be allowed")
	}

	// 3rd request should be blocked
	if limiter.Allow(ip) {
		t.Error("third request should be blocked")
	}
}

func TestRateLimiter_ZeroBurst(t *testing.T) {
	requestsPerMinute := 60
	burst := 0
	limiter := NewRateLimiter(requestsPerMinute, burst)
	defer limiter.Stop()

	ip := "192.168.1.30"

	// With zero burst, only refilled tokens should be available
	// First request should consume the initial token
	if !limiter.Allow(ip) {
		t.Error("first request should be allowed with zero burst")
	}

	// Second immediate request should be blocked
	if limiter.Allow(ip) {
		t.Error("second immediate request should be blocked with zero burst")
	}

	// Wait for token refill
	time.Sleep(1100 * time.Millisecond)

	// Should allow one more request
	if !limiter.Allow(ip) {
		t.Error("request after refill should be allowed")
	}
}

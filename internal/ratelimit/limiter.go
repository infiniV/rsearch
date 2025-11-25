// Package ratelimit implements per-IP rate limiting using the token bucket algorithm.
// It provides thread-safe rate limiting with automatic cleanup of stale entries.
package ratelimit

import (
	"sync"
	"time"
)

// tokenBucket represents a token bucket for a single IP address
type tokenBucket struct {
	tokens         float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

// RateLimiter implements per-IP rate limiting using token bucket algorithm
type RateLimiter struct {
	requestsPerMinute float64
	burst             float64
	tokensPerSecond   float64
	buckets           sync.Map // map[string]*tokenBucket
	cleanupInterval   time.Duration
	staleThreshold    time.Duration
	stopCleanup       chan struct{}
	cleanupWg         sync.WaitGroup
}

// NewRateLimiter creates a new rate limiter with specified requests per minute and burst size
func NewRateLimiter(requestsPerMinute int, burst int) *RateLimiter {
	return NewRateLimiterWithCleanup(requestsPerMinute, burst, 10*time.Minute, 1*time.Hour)
}

// NewRateLimiterWithCleanup creates a rate limiter with custom cleanup settings
func NewRateLimiterWithCleanup(requestsPerMinute int, burst int, cleanupInterval, staleThreshold time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMinute: float64(requestsPerMinute),
		burst:             float64(burst),
		tokensPerSecond:   float64(requestsPerMinute) / 60.0,
		cleanupInterval:   cleanupInterval,
		staleThreshold:    staleThreshold,
		stopCleanup:       make(chan struct{}),
	}

	// Start cleanup goroutine
	rl.cleanupWg.Add(1)
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	// Load or create bucket for this IP
	// Start with 1 token to allow the first request even with zero burst
	initialTokens := rl.burst
	if initialTokens < 1.0 {
		initialTokens = 1.0
	}
	value, _ := rl.buckets.LoadOrStore(ip, &tokenBucket{
		tokens:         initialTokens,
		lastRefillTime: time.Now(),
	})
	bucket := value.(*tokenBucket)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefillTime).Seconds()

	// Refill tokens based on elapsed time
	if elapsed > 0 {
		tokensToAdd := elapsed * rl.tokensPerSecond
		bucket.tokens += tokensToAdd
		// Cap at burst, but allow at least 1 token to accumulate
		maxTokens := rl.burst
		if maxTokens < 1.0 {
			maxTokens = 1.0
		}
		if bucket.tokens > maxTokens {
			bucket.tokens = maxTokens
		}
		bucket.lastRefillTime = now
	}

	// Check if we have at least one token
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}

// cleanupRoutine periodically removes stale bucket entries
func (rl *RateLimiter) cleanupRoutine() {
	defer rl.cleanupWg.Done()

	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes buckets that haven't been accessed in staleThreshold duration
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	staleIPs := []string{}

	rl.buckets.Range(func(key, value interface{}) bool {
		ip := key.(string)
		bucket := value.(*tokenBucket)

		bucket.mu.Lock()
		lastAccess := bucket.lastRefillTime
		bucket.mu.Unlock()

		if now.Sub(lastAccess) > rl.staleThreshold {
			staleIPs = append(staleIPs, ip)
		}

		return true
	})

	// Delete stale entries
	for _, ip := range staleIPs {
		rl.buckets.Delete(ip)
	}
}

// Stop gracefully shuts down the rate limiter and cleanup goroutine
func (rl *RateLimiter) Stop() {
	select {
	case <-rl.stopCleanup:
		// Already stopped
		return
	default:
		close(rl.stopCleanup)
		rl.cleanupWg.Wait()
	}
}

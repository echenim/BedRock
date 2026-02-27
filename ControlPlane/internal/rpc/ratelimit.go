package rpc

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter provides per-IP request rate limiting for the HTTP gateway (audit S3).
// Uses a simple token-bucket algorithm with periodic cleanup of stale entries.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // max requests per window
	window   time.Duration // refill window
	cleanup  time.Duration // how often to purge stale entries
	done     chan struct{}
}

type visitor struct {
	tokens    int
	lastSeen  time.Time
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter allowing `rate` requests per `window` per IP.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		cleanup:  2 * window,
		done:     make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Allow returns true if the request from the given IP is within the rate limit.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1,
			lastSeen:  now,
			lastReset: now,
		}
		return true
	}

	v.lastSeen = now

	// Refill tokens if the window has elapsed.
	if now.Sub(v.lastReset) >= rl.window {
		v.tokens = rl.rate
		v.lastReset = now
	}

	if v.tokens <= 0 {
		return false
	}

	v.tokens--
	return true
}

// Stop halts the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.cleanup)
			for ip, v := range rl.visitors {
				if v.lastSeen.Before(cutoff) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// extractIP extracts the client IP from the request, stripping the port.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (first entry is the client IP).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP only.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Fall back to RemoteAddr.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimitMiddleware wraps an http.Handler with per-IP rate limiting.
func RateLimitMiddleware(limiter *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !limiter.Allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

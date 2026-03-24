package server

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu          sync.Mutex
	count       int
	limit       int
	reset       time.Duration
	windowStart time.Time
}

func newRateLimiter(limit, resetSecs int) *rateLimiter {
	return &rateLimiter{
		limit:       limit,
		reset:       time.Duration(resetSecs) * time.Second,
		windowStart: time.Now(),
	}
}

// allow returns true if the request is within the rate limit.
// retryAfter is non-zero when a reset window is configured.
func (rl *rateLimiter) allow() (ok bool, retryAfter time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.reset > 0 && time.Since(rl.windowStart) >= rl.reset {
		rl.count = 0
		rl.windowStart = time.Now()
	}
	rl.count++
	if rl.count > rl.limit {
		if rl.reset > 0 {
			retryAfter = max(rl.reset-time.Since(rl.windowStart), 0)
		}
		return false, retryAfter
	}
	return true, 0
}

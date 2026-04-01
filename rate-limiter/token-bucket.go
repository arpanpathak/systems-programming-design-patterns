package rate_limiter

import (
	"sync"
	"time"
)

// TokenBucket implements a simple rate limiter.
type TokenBucket struct {
	rate       float64 // tokens per second
	capacity   float64 // max tokens
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}

// Example usage:
// limiter := NewTokenBucket(10, 20) // 10 tokens/sec, burst 20
// if limiter.Allow() { ... }

// TODO: Play with the data structure and implement a simple test case to see how it works.

package main

import (
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	rate        float64
	capacity    float64
	tokens      float64
	lastUpdated time.Time
}

func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		rate:        rate,
		capacity:    capacity,
		tokens:      capacity,
		lastUpdated: time.Now(),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastUpdated).Seconds()

	r.tokens += elapsed * r.rate
	if r.tokens > r.capacity {
		r.tokens = r.capacity
	}
	r.lastUpdated = now

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}

	return false
}

func main() {
	limiter := NewRateLimiter(2.0, 3.0)

	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			fmt.Printf("Request %d: Allowed\n", i+1)
		} else {
			fmt.Printf("Request %d: Denied\n", i+1)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

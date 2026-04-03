// Rate Limiter: Token Bucket algorithm.
package main

import (
	"context"
	"fmt"
	"time"
)

type TokenBucket struct {
	tokens     chan struct{}
	refillRate time.Duration
	done       chan struct{}
}

func NewTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	tb := &TokenBucket{
		tokens:     make(chan struct{}, capacity),
		refillRate: refillRate,
		done:       make(chan struct{}),
	}
	for i := 0; i < capacity; i++ {
		tb.tokens <- struct{}{}
	}
	go func() {
		ticker := time.NewTicker(refillRate)
		defer ticker.Stop()
		for {
			select {
			case <-tb.done:
				return
			case <-ticker.C:
				select {
				case tb.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()
	return tb
}

func (tb *TokenBucket) Allow() bool {
	select {
	case <-tb.tokens:
		return true
	default:
		return false
	}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	select {
	case <-tb.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (tb *TokenBucket) Stop() { close(tb.done) }

func main() {
	fmt.Println("=== Rate Limiter (Token Bucket) ===")

	limiter := NewTokenBucket(5, 50*time.Millisecond)
	defer limiter.Stop()

	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			fmt.Printf("  Request %d: ALLOWED\n", i)
		} else {
			fmt.Printf("  Request %d: RATE LIMITED\n", i)
		}
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println("  After waiting for refill:")
	for i := 10; i < 13; i++ {
		if limiter.Allow() {
			fmt.Printf("  Request %d: ALLOWED\n", i)
		}
	}
}

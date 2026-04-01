package resiliency

import (
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
)

// BulkheadIsolation mimics the ship architectural concept:
// If one compartment floods, the ship shouldn't sink.
// In Envoy proxying, if the "Auth Service" API is hanging, all 10,000 proxy Goroutines
// might get stuck waiting for it, starving the "Billing Service" of CPU resources!
//
// The Solution: A Weighted Semaphore. We limit the max concurrent requests allowed
// to hit ANY single upstream service. If the Auth Bulkhead is full, we fast-fail
// those requests, keeping the rest of the Proxy healthy!
func RunBulkheadIsolation() {
	fmt.Println("=== Bulkhead Isolation Pattern (Semaphore) ===")

	// 1. Create a Semaphore bounding this specific upstream to only 3 concurrent requests max!
	maxAuthConcurrentThreads := int64(3)
	authBulkhead := semaphore.NewWeighted(maxAuthConcurrentThreads)

	// 2. Simulate 10 incoming requests hitting the Auth Service simultaneously
	for i := 1; i <= 10; i++ {
		go func(id int) {
			// Proxy attempts to acquire a "compartment" space in the bulkhead.
			// TryAcquire doesn't block! It fast-fails immediately if the bulkhead is full.
			if !authBulkhead.TryAcquire(1) {
				fmt.Printf("[Request %d] Bulkhead FULL! Fast-failing HTTP 429 Too Many Requests.\n", id)
				return
			}

			// Defer releasing the semaphore token back to the proxy!
			defer authBulkhead.Release(1)

			fmt.Printf("[Request %d] Acquired bulkhead slot. Processing upstream request...\n", id)
			// Simulate long, hanging upstream Auth network call
			time.Sleep(1 * time.Second)
		}(i)
	}

	// This shows that only 3 will execute, and 7 will instantly fail,
	// saving the Proxy's memory/goroutine stack!
	time.Sleep(2 * time.Second)
}

// Notice: In Go systems programming, the Semaphore pattern is substantially better
// than spawning channels of empty structs `make(chan struct{}, 3)`. Semaphores
// via `x/sync/semaphore` allow context cancellations (`Acquire(ctx, 1)`) and weighted acquisitions!

// Semaphore pattern using buffered channel + context-based cancellation.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// --- Semaphore via buffered channel ---

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(max int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, max)}
}

func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

func main() {
	fmt.Println("=== Semaphore Pattern ===")
	sem := NewSemaphore(3)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			fmt.Printf("  Worker %d: acquired semaphore\n", id)
			time.Sleep(50 * time.Millisecond)
		}(i)
	}
	wg.Wait()
	fmt.Println("  All workers done")

	// --- Context cancellation ---
	fmt.Println("\n=== Context Cancellation ===")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("  Worker cancelled: %v\n", ctx.Err())
				return
			case <-ticker.C:
				fmt.Println("  Working...")
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	// WithTimeout
	fmt.Println("\n=== Context WithTimeout ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Work completed")
	case <-ctx2.Done():
		fmt.Printf("  Timed out: %v\n", ctx2.Err())
	}

	// WithValue
	type ctxKey string
	ctx3 := context.WithValue(context.Background(), ctxKey("requestID"), "req-12345")
	if v := ctx3.Value(ctxKey("requestID")); v != nil {
		fmt.Println("  Request ID:", v)
	}
}

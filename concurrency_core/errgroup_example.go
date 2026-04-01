package concurrency_core

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

// ErrGroupExample simulates fanning out proxy requests to multiple backends simultaneously.
// If ANY of the requests fail, the entire group is canceled via Context to save resources.
// This is exactly how Envoy handles parallel upstream requests with fault tolerance.
func ErrGroupExample() {
	fmt.Println("\n=== ErrGroup Example ===")

	// errgroup.WithContext creates a background group. If any goroutine returns an error,
	// 'ctx' gets canceled, signaling the surviving goroutines to abort early.
	g, ctx := errgroup.WithContext(context.Background())

	urls := []string{
		"http://backend-1.internal",
		"http://backend-2.internal",
		"http://backend-3.error", // We will simulate a failure here
	}

	for _, url := range urls {
		// Capture variable correctly in loop (though in Go >= 1.22 this isn't strictly necessary)
		u := url
		g.Go(func() error {
			return simulateProxyFetch(ctx, u)
		})
	}

	// Wait blocks until all functions finish OR one returns an error.
	if err := g.Wait(); err != nil {
		fmt.Printf("ErrGroup aborted successfully due to: %v\n", err)
	} else {
		fmt.Println("All requests succeeded!")
	}
}

func simulateProxyFetch(ctx context.Context, url string) error {
	select {
	case <-ctx.Done():
		// Another goroutine failed, so we should safely clean up and abort!
		fmt.Printf("Aborting fetch for %s due to cancellation\n", url)
		return ctx.Err()
	case <-time.After(50 * time.Millisecond): // Simulate network latency
		if url == "http://backend-3.error" {
			return fmt.Errorf("backend %s unreachable", url)
		}
		fmt.Printf("Successfully fetched %s\n", url)
		return nil
	}
}

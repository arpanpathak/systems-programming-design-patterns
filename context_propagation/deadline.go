package context_propagation

import (
	"context"
	"fmt"
	"time"
)

// RunDeadlineExample simulates how API gateways and Edge Proxies drop backing
// connections cleanly when a strict deadline passes, avoiding goroutine leaks.
func RunDeadlineExample() {
	fmt.Println("=== Context Deadline Example ===")

	// 1. Create a timeout context (SLA: maximum 100ms allowed)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	// Cancel must be called to release resources held by the context timer!
	defer cancel()

	fmt.Println("Proxy: Handing off request to backend...")

	// Channels act as our communication bus
	resultChan := make(chan string)

	go doBackendWork(ctx, resultChan)

	select {
	case res := <-resultChan:
		fmt.Printf("Proxy: Success [%s]\n", res)
	case <-ctx.Done():
		// ctx.Done() closes when the timeout is reached or cancel() is explicitly called
		fmt.Println("Proxy: SLA Enforced! Dropping connection (HTTP 504 Deadline Exceeded)")
	}
}

// doBackendWork acts as our upstream service resolving the request.
func doBackendWork(ctx context.Context, out chan<- string) {
	// Simulate slow db call taking 200ms (violates the 100ms SLA!)
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		// We finished work, hopefully proxy is still listening
		select {
		case out <- "200 OK":
		case <-ctx.Done():
			// The proxy has already stopped listening. We catch this here to
			// avoid blocking forever on sending to 'out'.
			fmt.Println("Backend: Proxy stopped listening. Rolling back transaction.")
		}
	case <-ctx.Done():
		// If the context is canceled before we finish work, we can abort the heavy computation early!
		fmt.Println("Backend: Context canceled early! Aborting network call immediately.")
	}
}

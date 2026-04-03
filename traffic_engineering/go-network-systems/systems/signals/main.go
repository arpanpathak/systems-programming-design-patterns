// Signal handling for graceful shutdown — critical for network services.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== Signal Handling ===")

	// In production: gracefully shut down on SIGINT/SIGTERM
	// This is critical for network services (drain connections, flush buffers)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		select {
		case sig := <-sigCh:
			fmt.Printf("  Received signal: %v\n", sig)
			// In real app: initiate graceful shutdown
		case <-ctx.Done():
			fmt.Println("  Signal demo timed out (no signal sent)")
		}
	}()

	<-ctx.Done()

	// Ignore signals
	signal.Reset(syscall.SIGINT)
	fmt.Println("  Signal handling demonstrated")
}

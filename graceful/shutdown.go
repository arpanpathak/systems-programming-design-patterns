package graceful

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdownServer demonstrates how edge traffic proxies guarantee no requests
// are dropped during deployments (SIGTERM) or ctrl+c interruptions (SIGINT).
func GracefulShutdownServer(address string) {
	// A basic mock HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Sleep simulates long-running connections (e.g. large file upload)
		time.Sleep(3 * time.Second)
		w.Write([]byte("Graceful Shutdown Handler Response\n"))
	})

	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	// 1. Kick off server in a separate Goroutine so we don't block
	go func() {
		log.Printf("Graceful HTTP Server listening on %s...\n", srv.Addr)
		// ErrServerClosed is expected when we call srv.Shutdown
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen Error: %v\n", err)
		}
	}()

	// 2. Setup standard OS Signal channels. Envoy waits over domain socket or SIGTERM.
	quit := make(chan os.Signal, 1)

	// `signal.Notify` registers the given channel to receive notifications of the specified signals.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block here until a Kill signal is caught!
	sig := <-quit
	log.Printf("OS Interruption Signal Caught: %+v. Initiating graceful shutdown sequence...\n", sig)

	// 3. Create context with timeout for graceful shutdown.
	// We allow inflight requests maximum 10 seconds to finish before forceful termination.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 4. `Shutdown` gracefully closes server, blocking until all connections are drained
	// or the context timeframe expires.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Force Shutdown Failed: %v\n", err)
	}

	log.Println("Graceful Shutdown completed successfully. Process exiting safely.")
}

// HTTP Server with middleware chain, SSE, and proper timeouts.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Middleware func(http.Handler) http.Handler

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		fmt.Printf("  [LOG] %s %s %d %v\n",
			r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func RateLimitMiddleware(rps int) Middleware {
	var mu sync.Mutex
	tokens := rps
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rps))
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			if tokens < rps {
				tokens++
			}
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			if tokens <= 0 {
				mu.Unlock()
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			tokens--
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("  [RECOVERY] panic: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func ChainMiddleware(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"headers": r.Header,
		"body":    string(body),
	})
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for i := 0; i < 5; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
			fmt.Fprintf(w, "data: Event %d at %s\n\n", i, time.Now().Format(time.RFC3339))
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func main() {
	fmt.Println("=== HTTP Server with Middleware ===")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/sse", sseHandler)

	handler := ChainMiddleware(mux, RecoveryMiddleware, LoggingMiddleware)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Listen error: %v\n", err)
		return
	}
	addr := listener.Addr().String()

	server := &http.Server{
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	go server.Serve(listener)
	defer server.Close()

	fmt.Printf("  HTTP server on %s\n", addr)

	client := &http.Client{Timeout: 10 * time.Second}

	// Health
	resp, _ := client.Get("http://" + addr + "/health")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("  GET /health: %s", body)

	// Echo
	resp, _ = client.Post("http://"+addr+"/echo", "application/json",
		strings.NewReader(`{"message":"hello"}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("  POST /echo: %s", body)

	// SSE
	fmt.Println("  SSE events:")
	resp, _ = client.Get("http://" + addr + "/sse")
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "data:") {
			fmt.Printf("    %s\n", line)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

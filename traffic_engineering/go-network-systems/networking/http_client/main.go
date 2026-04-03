// HTTP Client with connection pooling, retry, and backoff.
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func createHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

type RetryClient struct {
	client     *http.Client
	maxRetries int
	baseDelay  time.Duration
}

func NewRetryClient() *RetryClient {
	return &RetryClient{
		client:     createHTTPClient(),
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
	}
}

func (rc *RetryClient) Do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= rc.maxRetries; attempt++ {
		if attempt > 0 {
			delay := rc.baseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		resp, err := rc.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func main() {
	fmt.Println("=== HTTP Client with Retry ===")

	// Start a simple test server
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK\n"))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Listen error: %v\n", err)
		return
	}
	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	addr := listener.Addr().String()
	fmt.Printf("  Test server on %s\n", addr)

	rc := NewRetryClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := rc.Do(ctx, "GET", "http://"+addr+"/ok", nil)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Response: %s", body)

	fmt.Println("\n  Client best practices:")
	fmt.Println("  - Reuse *http.Client (connection pooling)")
	fmt.Println("  - Set timeouts (read, write, TLS, overall)")
	fmt.Println("  - Use context for cancellation")
	fmt.Println("  - Retry with exponential backoff")
	fmt.Println("  - Close response bodies")
}

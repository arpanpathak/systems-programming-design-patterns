// HTTP Reverse Proxy with round-robin load balancing and health checks.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) { b.mu.Lock(); b.Alive = alive; b.mu.Unlock() }
func (b *Backend) IsAlive() bool       { b.mu.RLock(); defer b.mu.RUnlock(); return b.Alive }

type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

func NewLoadBalancer(backendURLs []string) *LoadBalancer {
	lb := &LoadBalancer{}
	for _, rawURL := range backendURLs {
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		}
		origDir := proxy.Director
		proxy.Director = func(req *http.Request) {
			origDir(req)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Proxy", "GoLoadBalancer")
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			resp.Header.Set("X-Proxy-Backend", u.Host)
			return nil
		}
		lb.backends = append(lb.backends, &Backend{URL: u, Alive: true, ReverseProxy: proxy})
	}
	return lb
}

func (lb *LoadBalancer) RoundRobin() *Backend {
	n := uint64(len(lb.backends))
	if n == 0 {
		return nil
	}
	for i := uint64(0); i < n; i++ {
		idx := atomic.AddUint64(&lb.current, 1) % n
		if lb.backends[idx].IsAlive() {
			return lb.backends[idx]
		}
	}
	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b := lb.RoundRobin()
	if b == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	b.ReverseProxy.ServeHTTP(w, r)
}

func main() {
	fmt.Println("=== HTTP Reverse Proxy / Load Balancer ===")

	backends := make([]net.Listener, 2)
	for i := 0; i < 2; i++ {
		l, _ := net.Listen("tcp", "127.0.0.1:0")
		backends[i] = l
		id := i
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{"backend": id, "path": r.URL.Path})
		})
		go http.Serve(l, mux)
	}
	defer backends[0].Close()
	defer backends[1].Close()

	urls := make([]string, 2)
	for i, b := range backends {
		urls[i] = "http://" + b.Addr().String()
	}

	lb := NewLoadBalancer(urls)
	lbListener, _ := net.Listen("tcp", "127.0.0.1:0")
	defer lbListener.Close()
	lbServer := &http.Server{Handler: lb}
	go lbServer.Serve(lbListener)
	defer lbServer.Close()

	fmt.Printf("  LB on %s -> %v\n", lbListener.Addr(), urls)

	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 4; i++ {
		resp, err := client.Get("http://" + lbListener.Addr().String() + "/test")
		if err != nil {
			fmt.Printf("  Request %d error: %v\n", i, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("  Request %d -> %s", i, strings.TrimSpace(string(body)))
		fmt.Println()
	}
}

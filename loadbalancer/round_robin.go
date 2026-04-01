package loadbalancer

import (
	"net/url"
	"sync/atomic"
)

// RoundRobinLoadBalancer sequentially iterators over upstreams.
// Must be thread-safe as thousands of goroutine requests arrive concurrently.
type RoundRobinLoadBalancer struct {
	backends []*url.URL
	current  uint32 // Using uint32 to allow fast atomic CAS operations
}

func NewRoundRobinLoadBalancer(urls []string) *RoundRobinLoadBalancer {
	var backends []*url.URL
	for _, rawStr := range urls {
		parsed, _ := url.Parse(rawStr)
		backends = append(backends, parsed)
	}

	return &RoundRobinLoadBalancer{
		backends: backends,
		current:  0,
	}
}

// NextBackend retrieves the next load balanced node without locking a generic Mutex!
func (r *RoundRobinLoadBalancer) NextBackend() *url.URL {
	// Edge Case: No upstreams
	if len(r.backends) == 0 {
		return nil
	}

	// Atomic Add + Modulo operation.
	// Atomic additions do NOT require slow Mutex locking, crucial for Envoy-scale throughput!
	idx := atomic.AddUint32(&r.current, 1) % uint32(len(r.backends))
	return r.backends[idx]
}

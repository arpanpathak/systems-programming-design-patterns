package advanced_concurrency

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ProxyConfig represents an absolutely massive routing table (e.g., millions of URLs).
type ProxyConfig struct {
	Routes  map[string]string
	Version int
}

// Global state holding our active configuration
var currentProxyConfig atomic.Value

// RunLockFreeConfigSwapping simulates the Envoy xDS (Discovery Service) API.
// How do you update a routing table that millions of goroutines are currently reading from
// WITHOUT pausing them with a Mutex?
//
// Answer: You build the entirely new Config in the background, and then use
// `atomic.Value.Store()` to instantly swap the memory pointer!
func RunLockFreeConfigSwapping() {
	fmt.Println("=== Simulated xDS Lock-Free Routing Configuration Engine ===")

	// 1. Initial State Load
	initialConfig := &ProxyConfig{
		Version: 1,
		Routes:  map[string]string{"/api/v1": "backend-a.internal"},
	}
	currentProxyConfig.Store(initialConfig)

	var wg sync.WaitGroup

	// 2. Spawn 100 concurrent incoming API requests reading the config furiously!
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				// We load the pointer atomically. There are NO MUTEX LOCKS HERE!
				// Millions of threads can call this without bottlenecking.
				cfg := currentProxyConfig.Load().(*ProxyConfig)
				_ = cfg.Routes["/api/v1"] // Accessing the map is perfectly safe because nobody is modifying it!
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// 3. Simultaneously, a dynamic Config Push arrives from the Control Plane!
	go func() {
		time.Sleep(20 * time.Millisecond)

		fmt.Println("\n[Control Plane] Pushing massive Version 2 Route Table update...")
		// A. We create the entirely new table in isolated memory.
		newConfig := &ProxyConfig{
			Version: 2,
			Routes: map[string]string{
				"/api/v1": "backend-b.internal",
				"/api/v2": "backend-beta.internal",
			},
		}

		// B. The Magic: Atomically swap the memory pointer!
		// All subsequent reads by goroutines will instantly see Version 2.
		// Older goroutines currently processing Version 1 will finish safely because
		// we never mutated the actual V1 map, we just stopped pointing to it! (GC cleans V1 up later).
		currentProxyConfig.Store(newConfig)
		fmt.Println("[Control Plane] Atomic Ptr Swap complete! Zero dropped requests.")
	}()

	wg.Wait()

	finalCfg := currentProxyConfig.Load().(*ProxyConfig)
	fmt.Printf("\nFinal State Version: %d\n", finalCfg.Version)
}

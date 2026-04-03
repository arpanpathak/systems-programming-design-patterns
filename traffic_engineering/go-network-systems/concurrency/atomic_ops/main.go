// Atomic operations using Go 1.19+ typed atomics: atomic.Int64, atomic.Bool, etc.
// Also covers CAS, atomic.Value, and barrier pattern.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== Atomic Operations (typed atomics) ===")

	// atomic.Int64 — no pointer passing, cleaner API (Go 1.19+)
	var counter atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println("  Atomic counter:", counter.Load())

	// Compare and Swap (CAS) - foundation of lock-free algorithms
	var value atomic.Int64
	value.Store(10)
	swapped := value.CompareAndSwap(10, 20)
	fmt.Printf("  CAS: swapped=%v, value=%d\n", swapped, value.Load())

	swapped = value.CompareAndSwap(10, 30) // Won't swap, current is 20
	fmt.Printf("  CAS: swapped=%v, value=%d\n", swapped, value.Load())

	// atomic.Bool (Go 1.19+)
	var ready atomic.Bool
	ready.Store(true)
	fmt.Printf("  Atomic bool: ready=%v\n", ready.Load())

	// atomic.Uint64 (Go 1.19+)
	var seq atomic.Uint64
	seq.Add(1)
	seq.Add(1)
	fmt.Printf("  Atomic uint64: seq=%d\n", seq.Load())

	// atomic.Value for storing arbitrary types atomically
	var config atomic.Value
	config.Store(map[string]string{"host": "localhost", "port": "8080"})
	cfg := config.Load().(map[string]string)
	fmt.Printf("  Atomic config: host=%s, port=%s\n", cfg["host"], cfg["port"])

	// --- Barrier Pattern ---
	fmt.Println("\n=== Barrier Pattern ===")
	const numPhases = 3
	const numWorkers = 4

	for phase := 0; phase < numPhases; phase++ {
		var barrier sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			barrier.Add(1)
			go func(worker, p int) {
				defer barrier.Done()
				time.Sleep(time.Duration(worker*10) * time.Millisecond)
				fmt.Printf("  Worker %d completed phase %d\n", worker, p)
			}(w, phase)
		}
		barrier.Wait()
		fmt.Printf("  --- Phase %d barrier reached ---\n", phase)
	}
}

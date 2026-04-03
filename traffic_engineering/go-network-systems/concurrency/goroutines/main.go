// Basic Goroutines: lifecycle, WaitGroup, lightweight threads.
// Goroutines are M:N scheduled (multiplexed onto OS threads by Go runtime).
// They start with ~8KB stack that grows/shrinks dynamically.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Basic Goroutines ===")

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Goroutine %d running on logical thread\n", id)
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}(i)
	}

	wg.Wait()
	fmt.Println("  All goroutines completed")
}

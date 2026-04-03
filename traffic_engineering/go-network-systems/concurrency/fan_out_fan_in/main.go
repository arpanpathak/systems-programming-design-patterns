// Fan-Out / Fan-In pattern: distribute work across workers, collect results.
package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("=== Fan-Out / Fan-In ===")

	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// Fan-out: Multiple workers reading from the same channel
	numWorkers := 3
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				result := job * job
				fmt.Printf("  Worker %d processed job %d -> %d\n", id, job, result)
				results <- result
			}
		}(w)
	}

	// Send jobs
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs)

	// Fan-in: Collect all results
	go func() {
		wg.Wait()
		close(results)
	}()

	var total int
	for r := range results {
		total += r
	}
	fmt.Println("  Total sum of squares:", total)
}

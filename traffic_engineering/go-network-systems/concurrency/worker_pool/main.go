// Worker Pool: fixed number of workers processing jobs from a shared channel.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	ID      int
	Payload string
}

type Result struct {
	JobID  int
	Output string
	Err    error
}

func main() {
	fmt.Println("=== Worker Pool ===")

	const numWorkers = 4
	const numJobs = 20

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				results <- Result{
					JobID:  job.ID,
					Output: fmt.Sprintf("worker-%d processed '%s'", id, job.Payload),
				}
			}
		}(w)
	}

	for j := 0; j < numJobs; j++ {
		jobs <- Job{ID: j, Payload: fmt.Sprintf("task-%d", j)}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Printf("  Job %d: %s\n", r.JobID, r.Output)
	}
}

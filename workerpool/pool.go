package workerpool

import (
	"fmt"
	"sync"
	"time"
)

// Job represents a proxy work item, e.g., logging a request asynchronously.
type Job struct {
	ID    int
	Query string
}

// Dispatcher starts the worker pool and handles job submission smoothly.
type Dispatcher struct {
	JobQueue   chan Job
	numWorkers int
	wg         sync.WaitGroup
}

func NewDispatcher(numWorkers int, bufferSize int) *Dispatcher {
	return &Dispatcher{
		JobQueue:   make(chan Job, bufferSize),
		numWorkers: numWorkers,
	}
}

// Start boots up the worker goroutines. This is a common pattern to avoid spawning
// infinite goroutines which can lead to Out-Of-Memory errors on servers under load.
func (d *Dispatcher) Start() {
	fmt.Printf("Starting %d workers...\n", d.numWorkers)
	for i := 1; i <= d.numWorkers; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}
}

// worker constantly pulls from the JobQueue until the channel is closed.
func (d *Dispatcher) worker(id int) {
	defer d.wg.Done()
	for job := range d.JobQueue {
		// Simulate proxy workload
		fmt.Printf("Worker %d processing Job %d: %s\n", id, job.ID, job.Query)
		time.Sleep(10 * time.Millisecond) // Compute bound task
	}
	fmt.Printf("Worker %d shutting down.\n", id)
}

// Wait blocks until all jobs are done AND the pool gracefully shuts down.
func (d *Dispatcher) Wait() {
	close(d.JobQueue) // Signals to workers that no more jobs are coming
	d.wg.Wait()
}

// RunWorkerPoolExample demonstrates the pattern.
func RunWorkerPoolExample() {
	dispatcher := NewDispatcher(3, 100)
	dispatcher.Start()

	// Submit 10 async jobs
	for i := 1; i <= 10; i++ {
		dispatcher.JobQueue <- Job{ID: i, Query: fmt.Sprintf("SQL query %d", i)}
	}

	dispatcher.Wait()
	fmt.Println("Worker pool drained and closed.")
}

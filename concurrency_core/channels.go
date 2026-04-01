package concurrency_core

import (
	"fmt"
	"time"
)

// Demonstrates the fundamentals of Go channels (Critical for proxy traffic pipelines)
func ChannelsExample() {
	fmt.Println("=== Advanced Channels Example ===")

	// 1. Unbuffered Channel (Synchronous)
	// Requires both sender and receiver to be ready simultaneously.
	unbufChan := make(chan int)
	go func() {
		// Sender blocks until receiver takes it
		unbufChan <- 1
		fmt.Println("Sent to unbuffered channel")
	}()
	val := <-unbufChan
	fmt.Printf("Received from unbuffered channel: %d\n", val)

	// 2. Buffered Channel (Asynchronous up to capacity)
	// Allows bursty workloads (useful for batching proxy logs).
	bufChan := make(chan string, 2)
	bufChan <- "Log 1"
	bufChan <- "Log 2" // Won't block, buffer has space
	fmt.Printf("Received %s and %s from buffered chan\n", <-bufChan, <-bufChan)

	// 3. Directional Channels
	pipe := make(chan int, 5)
	producer(pipe)
	consumer(pipe)

	// 4. Select Statement (Multiplexing)
	// Key networking concept: Handling timeouts, cancellation, or concurrent ops.
	jobChan := make(chan string)
	timeout := time.After(50 * time.Millisecond) // Simulated SLA

	go func() {
		time.Sleep(100 * time.Millisecond) // Simulate slow backend
		jobChan <- "Done Job"
	}()

	select {
	case result := <-jobChan:
		fmt.Println("Success:", result)
	case <-timeout:
		fmt.Println("Timeout! Backend took too long (Proxy 504 Gateway Timeout)")
	}
}

// Unidirectional channel: Send-only
func producer(out chan<- int) {
	out <- 42
	close(out) // Sender should close the channel to prevent receiver deadlocks
}

// Unidirectional channel: Receive-only
func consumer(in <-chan int) {
	for val := range in {
		fmt.Println("Consumed:", val)
	}
}

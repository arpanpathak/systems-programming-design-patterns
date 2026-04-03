// Channels: unbuffered (synchronous), buffered (asynchronous), direction, range.
package main

import "fmt"

// send-only channel parameter
func producer(ch chan<- int, n int) {
	for i := 0; i < n; i++ {
		ch <- i * i
	}
	close(ch)
}

// receive-only channel parameter
func consumer(ch <-chan int) {
	for val := range ch {
		fmt.Printf("  Consumed: %d\n", val)
	}
}

func main() {
	// --- Unbuffered: sender blocks until receiver reads ---
	fmt.Println("=== Unbuffered Channels (Synchronous) ===")
	ch := make(chan string)
	go func() {
		ch <- "hello from goroutine"
	}()
	fmt.Println("  Received:", <-ch)

	// --- Buffered: sends don't block until buffer is full ---
	fmt.Println("\n=== Buffered Channels (Asynchronous) ===")
	bch := make(chan int, 3)
	bch <- 1
	bch <- 2
	bch <- 3
	fmt.Printf("  Channel length: %d, capacity: %d\n", len(bch), cap(bch))
	for i := 0; i < 3; i++ {
		fmt.Println("  Received:", <-bch)
	}

	// --- Channel direction: restrict to send-only or receive-only ---
	fmt.Println("\n=== Channel Direction ===")
	dirCh := make(chan int, 5)
	go producer(dirCh, 5)
	consumer(dirCh)
}

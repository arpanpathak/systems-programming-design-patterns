// Select patterns: multiplexing channels, non-blocking, timeout, done channel.
package main

import (
	"fmt"
	"time"
)

// selectStatement: blocks until one case can proceed.
// If multiple are ready, one is chosen at random (fair scheduling).
func selectStatement() {
	fmt.Println("=== Select Statement ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "channel 1"
	}()
	go func() {
		time.Sleep(30 * time.Millisecond)
		ch2 <- "channel 2"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println("  Received from ch1:", msg)
		case msg := <-ch2:
			fmt.Println("  Received from ch2:", msg)
		}
	}
}

// Non-blocking select uses `default` case.
func nonBlockingSelect() {
	fmt.Println("\n=== Non-blocking Select ===")

	ch := make(chan int, 1)

	select {
	case val := <-ch:
		fmt.Println("  Received:", val)
	default:
		fmt.Println("  No value ready (non-blocking)")
	}

	ch <- 42
	select {
	case ch <- 100:
		fmt.Println("  Sent 100")
	default:
		fmt.Println("  Channel full, skipped send")
	}
}

// Done channel: signal cancellation to goroutines via close().
func doneChannelPattern() {
	fmt.Println("\n=== Done Channel Pattern ===")

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		count := 0
		for {
			select {
			case <-done:
				fmt.Println("  Worker received done signal, exiting")
				return
			case <-ticker.C:
				count++
				fmt.Printf("  Working... iteration %d\n", count)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
	time.Sleep(50 * time.Millisecond)
}

// Timeout using time.After in select.
func timeoutPattern() {
	fmt.Println("\n=== Timeout Pattern ===")

	ch := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch <- "slow result"
	}()

	select {
	case result := <-ch:
		fmt.Println("  Got result:", result)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timed out waiting for result")
	}
}

func main() {
	selectStatement()
	nonBlockingSelect()
	doneChannelPattern()
	timeoutPattern()
}

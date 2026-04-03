// io.Pipe: synchronous in-memory pipe for streaming between goroutines.
package main

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== io.Pipe (In-process Pipe) ===")

	// io.Pipe creates a synchronous in-memory pipe.
	// Writes block until read — perfect for streaming data between goroutines.
	pr, pw := io.Pipe()

	var wg sync.WaitGroup

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer pw.Close()
		for i := 0; i < 5; i++ {
			fmt.Fprintf(pw, "message %d\n", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			fmt.Printf("  Pipe received: %s\n", scanner.Text())
		}
	}()

	wg.Wait()
}

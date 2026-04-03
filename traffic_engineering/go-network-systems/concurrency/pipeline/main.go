// Pipeline pattern: chain processing stages via channels.
// Or-Channel pattern: first-result-wins across multiple channels.
package main

import (
	"fmt"
	"time"
)

// --- Pipeline stages ---

func generator(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * 2
		}
		close(out)
	}()
	return out
}

// --- Or-Channel: returns when the first of N channels closes ---

func orChannel(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			mid := len(channels) / 2
			select {
			case <-orChannel(channels[:mid]...):
			case <-orChannel(channels[mid:]...):
			}
		}
	}()
	return orDone
}

func main() {
	fmt.Println("=== Pipeline Pattern ===")
	// Chain: generate -> square -> double
	ch := double(square(generator(1, 2, 3, 4, 5)))
	for v := range ch {
		fmt.Println("  Pipeline output:", v)
	}

	fmt.Println("\n=== Or-Channel (First Wins) ===")
	sig := func(after time.Duration) <-chan interface{} {
		ch := make(chan interface{})
		go func() {
			defer close(ch)
			time.Sleep(after)
		}()
		return ch
	}

	start := time.Now()
	<-orChannel(
		sig(2*time.Second),
		sig(5*time.Second),
		sig(100*time.Millisecond), // This one wins
		sig(1*time.Second),
	)
	fmt.Printf("  First signal received after %v\n", time.Since(start).Round(time.Millisecond))
}

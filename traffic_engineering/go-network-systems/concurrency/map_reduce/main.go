// Map-Reduce + Bounded Parallelism.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type WordCount map[string]int

func mapReduceDemo() {
	fmt.Println("=== Map-Reduce ===")

	data := []string{
		"hello world", "foo bar baz", "hello foo",
		"world bar", "baz hello world",
	}

	mapResults := make(chan WordCount, len(data))

	var wg sync.WaitGroup
	for _, chunk := range data {
		wg.Add(1)
		go func(text string) {
			defer wg.Done()
			counts := make(WordCount)
			start := 0
			for i := 0; i <= len(text); i++ {
				if i == len(text) || text[i] == ' ' {
					if i > start {
						counts[text[start:i]]++
					}
					start = i + 1
				}
			}
			mapResults <- counts
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(mapResults)
	}()

	finalCounts := make(WordCount)
	for partial := range mapResults {
		for word, count := range partial {
			finalCounts[word] += count
		}
	}

	fmt.Println("  Word counts:")
	for word, count := range finalCounts {
		fmt.Printf("    %s: %d\n", word, count)
	}
}

func boundedParallelismDemo() {
	fmt.Println("\n=== Bounded Parallelism ===")

	urls := []string{
		"url-1", "url-2", "url-3", "url-4", "url-5",
		"url-6", "url-7", "url-8", "url-9", "url-10",
	}

	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			fmt.Printf("  Fetched: %s\n", u)
		}(url)
	}
	wg.Wait()
}

func main() {
	mapReduceDemo()
	boundedParallelismDemo()
}

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	pingCh := make(chan struct{})
	pongCh := make(chan struct{})

	const MAX_ITERATIONS = 5

	// Ping goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < MAX_ITERATIONS; i++ {
			fmt.Println("Ping")
			pongCh <- struct{}{} // signal pong
			<-pingCh             // wait for next ping
			time.Sleep(1 * time.Second)
		}
	}()

	// Pong goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < MAX_ITERATIONS; i++ {
			<-pongCh // wait for ping
			fmt.Println("Pong")
			pingCh <- struct{}{} // signal ping back
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}

// Pub/Sub: topic-based message broadcasting to multiple subscribers.
package main

import (
	"fmt"
	"sync"
	"time"
)

type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan interface{}
	closed      bool
}

func NewPubSub() *PubSub {
	return &PubSub{subscribers: make(map[string][]chan interface{})}
}

func (ps *PubSub) Subscribe(topic string, bufSize int) <-chan interface{} {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan interface{}, bufSize)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic string, msg interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if ps.closed {
		return
	}
	for _, ch := range ps.subscribers[topic] {
		select {
		case ch <- msg:
		default: // back-pressure: drop if subscriber is slow
		}
	}
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.closed = true
	for _, subs := range ps.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
}

func main() {
	fmt.Println("=== Pub/Sub ===")

	ps := NewPubSub()
	sub1 := ps.Subscribe("events", 10)
	sub2 := ps.Subscribe("events", 10)
	sub3 := ps.Subscribe("alerts", 10)

	var wg sync.WaitGroup
	reader := func(name string, ch <-chan interface{}) {
		defer wg.Done()
		for msg := range ch {
			fmt.Printf("  %s received: %v\n", name, msg)
		}
	}

	wg.Add(3)
	go reader("sub1", sub1)
	go reader("sub2", sub2)
	go reader("sub3", sub3)

	ps.Publish("events", "user-login")
	ps.Publish("events", "page-view")
	ps.Publish("alerts", "cpu-high")

	time.Sleep(50 * time.Millisecond)
	ps.Close()
	wg.Wait()
}

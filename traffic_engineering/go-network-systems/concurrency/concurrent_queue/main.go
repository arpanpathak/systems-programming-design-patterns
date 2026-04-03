// Concurrent Queue with condition variables (bounded, blocking).
package main

import (
	"fmt"
	"sync"
)

type ConcurrentQueue struct {
	items    []interface{}
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	cap      int
	closed   bool
}

func NewConcurrentQueue(capacity int) *ConcurrentQueue {
	q := &ConcurrentQueue{
		items: make([]interface{}, 0, capacity),
		cap:   capacity,
	}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

func (q *ConcurrentQueue) Enqueue(item interface{}) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == q.cap && !q.closed {
		q.notFull.Wait()
	}
	if q.closed {
		return false
	}
	q.items = append(q.items, item)
	q.notEmpty.Signal()
	return true
}

func (q *ConcurrentQueue) Dequeue() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	if len(q.items) == 0 {
		return nil, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal()
	return item, true
}

func (q *ConcurrentQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}

func main() {
	fmt.Println("=== Concurrent Queue ===")

	q := NewConcurrentQueue(5)
	var wg sync.WaitGroup

	// Producers
	for p := 0; p < 3; p++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				item := fmt.Sprintf("P%d-item%d", id, i)
				q.Enqueue(item)
				fmt.Printf("  Produced: %s\n", item)
			}
		}(p)
	}

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 15; i++ {
			if item, ok := q.Dequeue(); ok {
				fmt.Printf("  Consumed: %v\n", item)
			}
		}
	}()

	wg.Wait()
	q.Close()
}

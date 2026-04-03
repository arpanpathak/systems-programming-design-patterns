// Sync primitives: sync.Once, sync.Cond, sync.Map, sync.Pool.
package main

import (
	"fmt"
	"sync"
)

// =============================================================================
// 1. sync.Once - Execute exactly once (singleton pattern)
// =============================================================================

type Singleton struct {
	Name string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance() *Singleton {
	once.Do(func() {
		fmt.Println("  Creating singleton instance")
		instance = &Singleton{Name: "the-one"}
	})
	return instance
}

// =============================================================================
// 2. sync.Cond - Condition Variable
// =============================================================================

type Queue struct {
	items []int
	cond  *sync.Cond
	mu    sync.Mutex
	cap   int
}

func NewQueue(capacity int) *Queue {
	q := &Queue{cap: capacity}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Enqueue(item int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == q.cap {
		q.cond.Wait() // Releases lock, waits, re-acquires
	}
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *Queue) Dequeue() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.cond.Signal()
	return item
}

// =============================================================================
// 3. sync.Map - Concurrent map (optimized for read-heavy or disjoint writes)
// =============================================================================

func syncMapDemo() {
	fmt.Println("\n=== sync.Map ===")

	var m sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id%3)
			m.Store(key, id)
		}(i)
	}
	wg.Wait()

	m.Range(func(key, value interface{}) bool {
		fmt.Printf("  %s: %v\n", key, value)
		return true
	})

	actual, loaded := m.LoadOrStore("new-key", "new-value")
	fmt.Printf("  LoadOrStore: actual=%v, loaded=%v\n", actual, loaded)
}

// =============================================================================
// 4. sync.Pool - Object pool to reduce GC pressure
// =============================================================================

func syncPoolDemo() {
	fmt.Println("\n=== sync.Pool ===")

	pool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("  Creating new buffer")
			return make([]byte, 1024)
		},
	}

	buf := pool.Get().([]byte)
	fmt.Printf("  Got buffer of size %d\n", len(buf))
	pool.Put(buf) // Return for reuse

	buf2 := pool.Get().([]byte)
	fmt.Printf("  Got buffer of size %d (reused)\n", len(buf2))
}

func main() {
	// sync.Once
	fmt.Println("=== sync.Once ===")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s := GetInstance()
			fmt.Printf("  Goroutine %d got: %s\n", id, s.Name)
		}(i)
	}
	wg.Wait()

	// sync.Cond
	fmt.Println("\n=== sync.Cond (Condition Variable) ===")
	q := NewQueue(3)
	go func() {
		for i := 0; i < 10; i++ {
			q.Enqueue(i)
			fmt.Printf("  Produced: %d\n", i)
		}
	}()
	for i := 0; i < 10; i++ {
		val := q.Dequeue()
		fmt.Printf("  Consumed: %d\n", val)
	}

	// sync.Map
	syncMapDemo()

	// sync.Pool
	syncPoolDemo()
}

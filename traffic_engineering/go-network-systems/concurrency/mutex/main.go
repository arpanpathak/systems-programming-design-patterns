// Mutex and RWMutex: mutual exclusion for shared state.
package main

import (
	"fmt"
	"sync"
)

// SafeCounter uses Mutex for exclusive access.
type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v[key]++
}

func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v[key]
}

// Cache uses RWMutex: multiple concurrent readers, exclusive writer.
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func main() {
	fmt.Println("=== Mutex ===")
	c := SafeCounter{v: make(map[string]int)}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc("key")
		}()
	}
	wg.Wait()
	fmt.Println("  Final count:", c.Value("key")) // Always 1000

	fmt.Println("\n=== RWMutex ===")
	cache := &Cache{data: make(map[string]string)}
	cache.Set("greeting", "hello")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if val, ok := cache.Get("greeting"); ok {
				fmt.Printf("  Reader %d: %s\n", id, val)
			}
		}(i)
	}
	wg.Wait()
}

// Concurrent LRU Cache with mutex protection.
package main

import (
	"container/list"
	"fmt"
	"sync"
)

type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type lruEntry struct {
	key   string
	value interface{}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*lruEntry).value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*lruEntry).value = value
		return
	}
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*lruEntry).key)
		}
	}
	entry := &lruEntry{key: key, value: value}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

func printLRU(c *LRUCache) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for e := c.order.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*lruEntry)
		fmt.Printf("%s=%v ", entry.key, entry.value)
	}
	fmt.Println()
}

func main() {
	fmt.Println("=== Concurrent LRU Cache ===")

	cache := NewLRUCache(3)
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)
	fmt.Printf("  After adding a,b,c: ")
	printLRU(cache)

	cache.Get("a")    // makes 'a' most recent
	cache.Put("d", 4) // evicts 'b'
	fmt.Printf("  After Get(a), Put(d): ")
	printLRU(cache)

	if _, ok := cache.Get("b"); !ok {
		fmt.Println("  'b' was evicted (correct)")
	}

	// Concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id%5)
			cache.Put(key, id)
			cache.Get(key)
		}(i)
	}
	wg.Wait()
	fmt.Println("  Concurrent access completed successfully")
}

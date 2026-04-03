// Consistent hashing for proxy routing — minimizes key redistribution.
package main

import (
	"fmt"
	"sync"
)

type ConsistentHash struct {
	ring     map[uint32]string
	sorted   []uint32
	replicas int
	mu       sync.RWMutex
}

func NewConsistentHash(replicas int) *ConsistentHash {
	return &ConsistentHash{ring: make(map[uint32]string), replicas: replicas}
}

func (ch *ConsistentHash) hash(key string) uint32 {
	var h uint32 = 2166136261 // FNV-1a
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

func (ch *ConsistentHash) Add(server string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for i := 0; i < ch.replicas; i++ {
		h := ch.hash(fmt.Sprintf("%s-%d", server, i))
		ch.ring[h] = server
		ch.sorted = append(ch.sorted, h)
	}
	// Insertion sort
	for i := 1; i < len(ch.sorted); i++ {
		key := ch.sorted[i]
		j := i - 1
		for j >= 0 && ch.sorted[j] > key {
			ch.sorted[j+1] = ch.sorted[j]
			j--
		}
		ch.sorted[j+1] = key
	}
}

func (ch *ConsistentHash) Get(key string) string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	if len(ch.sorted) == 0 {
		return ""
	}
	h := ch.hash(key)
	lo, hi := 0, len(ch.sorted)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if ch.sorted[mid] < h {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if ch.sorted[lo] < h {
		lo = 0
	}
	return ch.ring[ch.sorted[lo]]
}

func main() {
	fmt.Println("=== Consistent Hashing ===")

	ch := NewConsistentHash(100) // 100 virtual nodes per server
	ch.Add("server-A")
	ch.Add("server-B")
	ch.Add("server-C")

	keys := []string{"user-1", "user-2", "user-3", "user-4", "session-abc", "request-xyz"}

	fmt.Println("  Initial routing:")
	for _, key := range keys {
		fmt.Printf("    %s -> %s\n", key, ch.Get(key))
	}

	ch.Add("server-D")
	fmt.Println("  After adding server-D:")
	for _, key := range keys {
		fmt.Printf("    %s -> %s\n", key, ch.Get(key))
	}

	fmt.Println("\n  Key property: adding server-D only remaps ~1/N keys")
}

// Sharded Map: reduces lock contention by distributing keys across shards.
package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

const numShards = 16

type ShardedMap struct {
	shards [numShards]*mapShard
}

type mapShard struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func NewShardedMap() *ShardedMap {
	sm := &ShardedMap{}
	for i := 0; i < numShards; i++ {
		sm.shards[i] = &mapShard{items: make(map[string]interface{})}
	}
	return sm
}

func (sm *ShardedMap) getShard(key string) *mapShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return sm.shards[h.Sum32()%numShards]
}

func (sm *ShardedMap) Set(key string, value interface{}) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = value
}

func (sm *ShardedMap) Get(key string) (interface{}, bool) {
	shard := sm.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, ok := shard.items[key]
	return val, ok
}

func (sm *ShardedMap) Delete(key string) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

func (sm *ShardedMap) Len() int {
	total := 0
	for _, shard := range sm.shards {
		shard.mu.RLock()
		total += len(shard.items)
		shard.mu.RUnlock()
	}
	return total
}

func main() {
	fmt.Println("=== Sharded Map ===")

	sm := NewShardedMap()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			sm.Set(key, id*id)
		}(i)
	}
	wg.Wait()

	fmt.Printf("  Sharded map size: %d\n", sm.Len())
	if val, ok := sm.Get("key-42"); ok {
		fmt.Printf("  key-42 = %v\n", val)
	}
}

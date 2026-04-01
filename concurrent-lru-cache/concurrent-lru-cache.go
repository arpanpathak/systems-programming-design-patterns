package concurrent_lru_cache

import (
	"container/list"
	"hash/crc32"
	"sync"
)

// The Problem: A standard map + sync.Mutex becomes a massive bottleneck
// when an Edge Proxy receives 50,000 requests per second. Every goroutine blocks waiting
// for the global lock just to read from the cache!
//
// The Solution: Sharding. We split the cache into smaller, independent buckets.
// If we have 256 shards, the statistical probability of two goroutines hitting the
// exact same mutex lock drops dramatically. (Envoy/Redis heavily use this pattern).

// CacheItem holds the actual Key and Value as a standard struct.
// (Storing Key is important so we can delete from the map when evicting the tail of the LRU).
type CacheItem struct {
	Key   string
	Value any // Generic value payload
}

// LRUShard manages a fraction of the total cache data.
// It uses a standard lock, but it only locks for the few keys mapped specifically to this shard!
type LRUShard struct {
	mu        sync.RWMutex
	capacity  int
	items     map[string]*list.Element
	evictList *list.List
}

// ShardedLRUCache is the container holding all our parallel shards.
type ShardedLRUCache struct {
	shards    []*LRUShard
	numShards uint32 // Using uint32 for fast modulo arithmetic without type casting
}

// NewShardedLRUCache initializes the parallel caching structure.
// 'totalCapacity' is split evenly across all shards.
func NewShardedLRUCache(numShards int, totalCapacity int) *ShardedLRUCache {
	shardCapacity := totalCapacity / numShards
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	c := &ShardedLRUCache{
		shards:    make([]*LRUShard, numShards),
		numShards: uint32(numShards),
	}

	for i := 0; i < numShards; i++ {
		c.shards[i] = &LRUShard{
			capacity:  shardCapacity,
			items:     make(map[string]*list.Element),
			evictList: list.New(),
		}
	}
	return c
}

// getShard mathematically determinates which lock bucket a key belongs to.
// We use CRC32 because it's extremely fast and provides decent uniform distribution.
func (c *ShardedLRUCache) getShard(key string) *LRUShard {
	hash := crc32.ChecksumIEEE([]byte(key))
	return c.shards[hash%c.numShards]
}

// Put inserts or updates a key. Only the specific shard's Mutex is locked.
func (c *ShardedLRUCache) Put(key string, value any) {
	shard := c.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// 1. Update existing key: Move to front of LRU queue
	if ent, ok := shard.items[key]; ok {
		shard.evictList.MoveToFront(ent)
		ent.Value.(*CacheItem).Value = value
		return
	}

	// 2. Add New Key: Enforce capacity constraints
	if shard.evictList.Len() >= shard.capacity {
		c.removeOldest(shard)
	}

	// 3. Insert and track
	ent := shard.evictList.PushFront(&CacheItem{Key: key, Value: value})
	shard.items[key] = ent
}

// Get retrieves a key. RWMutex allows thousands of parallel READS if no writes are happening.
func (c *ShardedLRUCache) Get(key string) (any, bool) {
	shard := c.getShard(key)

	// We MUST acquire a write lock if we are updating the LRU list order.
	// (For strict performance, some architectures allow a small race condition on reads to avoid locking entirely).
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if ent, ok := shard.items[key]; ok {
		// Moving to front marks it as "Recently Used"
		shard.evictList.MoveToFront(ent)
		return ent.Value.(*CacheItem).Value, true
	}

	return nil, false
}

// removeOldest drops the least recently used item (the tail of the doubly-linked list).
// Note: This expects the shard.mu to ALREADY be locked by the caller!
func (c *ShardedLRUCache) removeOldest(shard *LRUShard) {
	ent := shard.evictList.Back()
	if ent != nil {
		shard.evictList.Remove(ent)
		kv := ent.Value.(*CacheItem)
		delete(shard.items, kv.Key) // Free memory from map
	}
}

package loadbalancer

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHashRing creates a hashed mapping of servers to uniformly map sticky requests
// across available backend nodes, avoiding re-sharding massive amounts of session data
// when a single upstream server goes down.
type ConsistentHashRing struct {
	mu       sync.RWMutex
	keys     []uint32          // Sorted array of hashed ring positions
	hashMap  map[uint32]string // Maps node hash back to the node's identifier URL
	replicas int               // Virtual Nodes (vNodes): to ensure better distribution on the ring
}

func NewConsistentHashRing(replicas int) *ConsistentHashRing {
	return &ConsistentHashRing{
		hashMap:  make(map[uint32]string),
		replicas: replicas,
	}
}

// AddNode registers a new upstream backend to the hash ring.
func (c *ConsistentHashRing) AddNode(node string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.replicas; i++ {
		// Create a Virtual Node (NodeName + Index) for better dispersion
		vn := strconv.Itoa(i) + node
		hash := c.hashFunc([]byte(vn))

		// Insert into the ring and track its mapping
		c.keys = append(c.keys, hash)
		c.hashMap[hash] = node
	}

	// Keys must be in ascending order so we can binary-search on the ring!
	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})
}

// GetNode routes a client (e.g., mapped by Client IP or Session ID) to the appropriate backend.
func (c *ConsistentHashRing) GetNode(clientIP string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.keys) == 0 {
		return ""
	}

	hash := c.hashFunc([]byte(clientIP))

	// Binary search to find the nearest clockwise Virtual Node (vNode)
	// whose hash is greater than or equal to the incoming client hash.
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	// If we wrap around the ring's 360 degrees boundary, we route back to the first node (index 0).
	if idx == len(c.keys) {
		idx = 0
	}

	return c.hashMap[c.keys[idx]]
}

// hashFunc calculates a 32-bit checksum hash.
// CRC32 is extremely fast making it acceptable for routing throughputs,
// though MD5/SHA256 handles clustering balance slightly better.
func (c *ConsistentHashRing) hashFunc(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

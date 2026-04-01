package thread_safe_queue

import (
	"errors"
	"sync/atomic"
)

// RingBuffer is a highly performant, lock-free (or low-lock) circular queue.
// In Traffic Engineering (Envoy/eBPF/XDP), allocating new slice memory dynamically
// causes Garbage Collection (GC) pauses which spike P99 latency.
// A Ring Buffer allocates memory exactly ONCE on boot, and simply overwrites
// old memory mathematically. This is the cornerstone of packet-processing queues!
type RingBuffer struct {
	buffer []any  // The fixed-size array holding our network packets/requests
	size   uint64 // Mathematical size constraint (Must be a power of 2 for bitwise optimization)
	head   uint64 // Atomic index where the Consumer reads (reads pull from Head)
	tail   uint64 // Atomic index where the Producer writes (writes push to Tail)
}

var ErrBufferFull = errors.New("ring buffer is at maximum capacity (Producer must drop packet)")
var ErrBufferEmpty = errors.New("ring buffer is completely empty (Consumer must wait)")

// NewRingBuffer allocates the queue exactly once.
func NewRingBuffer(size uint64) *RingBuffer {
	return &RingBuffer{
		buffer: make([]any, size),
		size:   size,
	}
}

// Push adds a new item to the queue. If it's full, we reject it (Backpressure).
func (rb *RingBuffer) Push(item any) error {
	// 1. Read the current atomic indices
	currentTail := atomic.LoadUint64(&rb.tail)
	currentHead := atomic.LoadUint64(&rb.head)

	// 2. Is it full? (If the distance between Head and Tail equals our Size constraints)
	if currentTail-currentHead >= rb.size {
		return ErrBufferFull // The producer is firing too fast!
	}

	// 3. Mathematical Modulo to wrap around the slice!
	// (Example: if size is 10, and tail is 11, index is 1).
	idx := currentTail % rb.size

	// 4. Write data to the pre-allocated index (Zero new memory allocations overhead!)
	rb.buffer[idx] = item

	// 5. Commit the new tracking index to memory atomically so Consumers can see it.
	atomic.AddUint64(&rb.tail, 1)

	return nil
}

// Pop retrieves the oldest item from the queue without deleting it from RAM.
func (rb *RingBuffer) Pop() (any, error) {
	// 1. Where are we currently?
	currentHead := atomic.LoadUint64(&rb.head)
	currentTail := atomic.LoadUint64(&rb.tail)

	// 2. Are we empty? (If Head caught up to Tail, there is no unread data!)
	if currentHead == currentTail {
		return nil, ErrBufferEmpty // Consumer is too fast!
	}

	// 3. Modulo math to locate the item
	idx := currentHead % rb.size

	// 4. Retrieve data (Notice we don't 'delete' it. The next time the Tail loops
	// around, it will simply overwrite this RAM sector).
	item := rb.buffer[idx]

	// 5. Erase reference to allow garbage collection if 'item' holds complex pointers
	rb.buffer[idx] = nil

	// 6. Commit the updated read pointer so the Producer knows it has free space.
	atomic.AddUint64(&rb.head, 1)

	return item, nil
}

// Lock-free SPSC (Single-Producer Single-Consumer) Ring Buffer.
package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type SPSCRingBuffer struct {
	buffer []interface{}
	cap    uint64
	head   uint64 // written by producer
	tail   uint64 // written by consumer
}

func NewSPSCRingBuffer(capacity uint64) *SPSCRingBuffer {
	return &SPSCRingBuffer{
		buffer: make([]interface{}, capacity),
		cap:    capacity,
	}
}

func (rb *SPSCRingBuffer) Enqueue(item interface{}) bool {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)
	if head-tail >= rb.cap {
		return false
	}
	rb.buffer[head%rb.cap] = item
	atomic.StoreUint64(&rb.head, head+1)
	return true
}

func (rb *SPSCRingBuffer) Dequeue() (interface{}, bool) {
	tail := atomic.LoadUint64(&rb.tail)
	head := atomic.LoadUint64(&rb.head)
	if tail >= head {
		return nil, false
	}
	item := rb.buffer[tail%rb.cap]
	atomic.StoreUint64(&rb.tail, tail+1)
	return item, true
}

func main() {
	fmt.Println("=== Lock-Free Ring Buffer (SPSC) ===")

	rb := NewSPSCRingBuffer(8)
	done := make(chan struct{})

	// Producer
	go func() {
		for i := 0; i < 20; i++ {
			for !rb.Enqueue(i) {
				time.Sleep(time.Millisecond)
			}
		}
		close(done)
	}()

	// Consumer
	count := 0
	for count < 20 {
		if val, ok := rb.Dequeue(); ok {
			fmt.Printf("  Dequeued: %v\n", val)
			count++
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	<-done
}

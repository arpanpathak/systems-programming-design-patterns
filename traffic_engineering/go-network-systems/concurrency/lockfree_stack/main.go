// Lock-Free Stack (Treiber Stack) using CAS (Compare-And-Swap).
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type lfNode struct {
	value interface{}
	next  unsafe.Pointer
}

type LockFreeStack struct {
	top unsafe.Pointer
	len int64
}

func (s *LockFreeStack) Push(val interface{}) {
	node := &lfNode{value: val}
	for {
		oldTop := atomic.LoadPointer(&s.top)
		node.next = oldTop
		if atomic.CompareAndSwapPointer(&s.top, oldTop, unsafe.Pointer(node)) {
			atomic.AddInt64(&s.len, 1)
			return
		}
	}
}

func (s *LockFreeStack) Pop() (interface{}, bool) {
	for {
		oldTop := atomic.LoadPointer(&s.top)
		if oldTop == nil {
			return nil, false
		}
		node := (*lfNode)(oldTop)
		newTop := atomic.LoadPointer(&node.next)
		if atomic.CompareAndSwapPointer(&s.top, oldTop, newTop) {
			atomic.AddInt64(&s.len, -1)
			return node.value, true
		}
	}
}

func (s *LockFreeStack) Len() int64 {
	return atomic.LoadInt64(&s.len)
}

func main() {
	fmt.Println("=== Lock-Free Stack (Treiber Stack) ===")

	stack := &LockFreeStack{}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			stack.Push(v)
		}(i)
	}
	wg.Wait()
	fmt.Printf("  Stack size after 100 pushes: %d\n", stack.Len())

	popped := int64(0)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := stack.Pop(); ok {
				atomic.AddInt64(&popped, 1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("  Popped %d items, stack size: %d\n", popped, stack.Len())
}

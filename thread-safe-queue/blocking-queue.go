package thread_safe_queue

import (
	"sync"
)

type SafeQueue struct {
	items []interface{}
	mu    sync.Mutex
	cond  *sync.Cond
}

func NewSafeQueue() *SafeQueue {
	q := &SafeQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *SafeQueue) Push(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *SafeQueue) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// Work-Stealing Deque: each worker has a local deque, idle workers steal from others.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type WorkStealingDeque struct {
	mu    sync.Mutex
	items []func() int
}

func (d *WorkStealingDeque) PushBottom(task func() int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items = append(d.items, task)
}

func (d *WorkStealingDeque) PopBottom() (func() int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.items) == 0 {
		return nil, false
	}
	n := len(d.items) - 1
	task := d.items[n]
	d.items = d.items[:n]
	return task, true
}

func (d *WorkStealingDeque) StealTop() (func() int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.items) == 0 {
		return nil, false
	}
	task := d.items[0]
	d.items = d.items[1:]
	return task, true
}

func (d *WorkStealingDeque) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.items)
}

func main() {
	fmt.Println("=== Work-Stealing Deque ===")

	const numWorkers = 4
	deques := make([]*WorkStealingDeque, numWorkers)
	for i := range deques {
		deques[i] = &WorkStealingDeque{}
	}

	// Load imbalanced work: worker 0 gets most tasks
	for i := 0; i < 40; i++ {
		id := i
		deques[0].PushBottom(func() int {
			time.Sleep(time.Millisecond)
			return id
		})
	}

	var totalProcessed atomic.Int64
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var processed int64

			for {
				// Try own deque first
				if task, ok := deques[workerID].PopBottom(); ok {
					task()
					processed++
					continue
				}
				// Try stealing from a random other worker
				stolen := false
				victim := rand.Intn(numWorkers)
				if victim != workerID {
					if task, ok := deques[victim].StealTop(); ok {
						task()
						processed++
						stolen = true
					}
				}
				if !stolen {
					// Check if any work remains
					anyWork := false
					for i := 0; i < numWorkers; i++ {
						if deques[i].Len() > 0 {
							anyWork = true
							break
						}
					}
					if !anyWork {
						break
					}
					time.Sleep(time.Millisecond)
				}
			}
			totalProcessed.Add(processed)
			fmt.Printf("  Worker %d processed %d tasks\n", workerID, processed)
		}(w)
	}

	wg.Wait()
	fmt.Printf("  Total processed: %d\n", totalProcessed.Load())
}

// Read-Copy-Update (RCU) pattern: readers access data lock-free via atomic pointer,
// writers create a new copy, swap pointer, and wait for readers to drain.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Config struct {
	Version int
	Setting string
}

type RCUConfig struct {
	ptr unsafe.Pointer // *Config
}

func NewRCUConfig(cfg *Config) *RCUConfig {
	r := &RCUConfig{}
	atomic.StorePointer(&r.ptr, unsafe.Pointer(cfg))
	return r
}

func (r *RCUConfig) Read() *Config {
	return (*Config)(atomic.LoadPointer(&r.ptr))
}

func (r *RCUConfig) Update(newCfg *Config) {
	atomic.StorePointer(&r.ptr, unsafe.Pointer(newCfg))
	// Grace period: wait for in-flight readers to finish
	time.Sleep(10 * time.Millisecond)
}

func main() {
	fmt.Println("=== Read-Copy-Update (RCU) Pattern ===")

	rcu := NewRCUConfig(&Config{Version: 1, Setting: "initial"})
	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				cfg := rcu.Read()
				fmt.Printf("  Reader %d: version=%d setting=%s\n", id, cfg.Version, cfg.Setting)
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	// Writer updates config a few times
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := 2; v <= 4; v++ {
			time.Sleep(20 * time.Millisecond)
			newCfg := &Config{Version: v, Setting: fmt.Sprintf("setting-v%d", v)}
			rcu.Update(newCfg)
			fmt.Printf("  Writer: updated to version %d\n", v)
		}
	}()

	wg.Wait()
	final := rcu.Read()
	fmt.Printf("  Final config: version=%d setting=%s\n", final.Version, final.Setting)
}

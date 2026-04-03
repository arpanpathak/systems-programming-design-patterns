// Bandwidth throttler using token bucket algorithm.
package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type Throttler struct {
	bytesPerSecond int64
	tokens         int64
	mu             sync.Mutex
	done           chan struct{}
}

func NewThrottler(bytesPerSecond int64) *Throttler {
	t := &Throttler{
		bytesPerSecond: bytesPerSecond,
		tokens:         bytesPerSecond,
		done:           make(chan struct{}),
	}
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		refill := bytesPerSecond / 10
		for {
			select {
			case <-t.done:
				return
			case <-ticker.C:
				t.mu.Lock()
				t.tokens += refill
				if t.tokens > bytesPerSecond {
					t.tokens = bytesPerSecond
				}
				t.mu.Unlock()
			}
		}
	}()
	return t
}

func (t *Throttler) ThrottledCopy(dst io.Writer, src io.Reader) (int64, error) {
	var total int64
	buf := make([]byte, 4096)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			for {
				t.mu.Lock()
				if t.tokens >= int64(n) {
					t.tokens -= int64(n)
					t.mu.Unlock()
					break
				}
				t.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
			written, writeErr := dst.Write(buf[:n])
			total += int64(written)
			if writeErr != nil {
				return total, writeErr
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return total, nil
			}
			return total, readErr
		}
	}
}

func (t *Throttler) Stop() {
	close(t.done)
}

func main() {
	fmt.Println("=== Bandwidth Throttler ===")

	throttler := NewThrottler(1024) // 1 KB/s
	defer throttler.Stop()

	src := strings.NewReader(strings.Repeat("x", 512))
	var dst strings.Builder

	start := time.Now()
	n, err := throttler.ThrottledCopy(&dst, src)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Copied: %d bytes in %v (target: 1024 B/s)\n", n, elapsed)
}

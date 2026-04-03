// sync.Pool for object reuse and GC insights.
package main

import (
	"fmt"
	"runtime"
	"sync"
)

type Buffer struct {
	data []byte
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{data: make([]byte, 0, 4096)}
	},
}

func getBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

func putBuffer(buf *Buffer) {
	buf.data = buf.data[:0]
	bufferPool.Put(buf)
}

func syncPoolDemo() {
	fmt.Println("\n--- sync.Pool for Object Reuse ---")

	var memBefore, memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Without pool: allocate new buffers each time
	for i := 0; i < 10000; i++ {
		buf := make([]byte, 4096)
		buf[0] = byte(i)
		_ = buf
	}

	runtime.GC()
	runtime.ReadMemStats(&memAfter)
	fmt.Printf("  Without pool - Allocs: %d\n", memAfter.TotalAlloc-memBefore.TotalAlloc)

	runtime.ReadMemStats(&memBefore)

	// With pool: reuse buffers
	for i := 0; i < 10000; i++ {
		buf := getBuffer()
		buf.data = append(buf.data, byte(i))
		putBuffer(buf)
	}

	runtime.GC()
	runtime.ReadMemStats(&memAfter)
	fmt.Printf("  With pool - Allocs: %d\n", memAfter.TotalAlloc-memBefore.TotalAlloc)
}

func gcDemo() {
	fmt.Println("\n--- Garbage Collector ---")

	// Go uses concurrent tri-color mark-and-sweep GC.
	// Three colors: white (unmarked), grey (marked, children not scanned), black (done).
	// Write barrier ensures correctness during concurrent marking.

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("  GC Cycles: %d\n", stats.NumGC)
	fmt.Printf("  Last GC pause: %d ns\n", stats.PauseNs[(stats.NumGC+255)%256])
	fmt.Printf("  GC CPU fraction: %.4f\n", stats.GCCPUFraction)

	runtime.GC()
	runtime.ReadMemStats(&stats)
	fmt.Printf("  After forced GC - cycles: %d\n", stats.NumGC)

	fmt.Println("  GOGC=100 (default): GC when heap grows 100%")
	fmt.Println("  Use GOGC=off + GOMEMLIMIT for memory-constrained environments")
}

type Resource struct {
	Name string
}

func finalizerDemo() {
	fmt.Println("\n--- Finalizers ---")

	// Finalizers run when GC collects an object.
	// WARNING: Not guaranteed to run on program exit.
	// Prefer explicit Close() methods.
	r := &Resource{Name: "db-connection"}
	runtime.SetFinalizer(r, func(res *Resource) {
		fmt.Printf("  Finalizer: cleaning up %s\n", res.Name)
	})

	fmt.Printf("  Resource created: %s\n", r.Name)
	r = nil
	runtime.GC()
	runtime.Gosched()
	fmt.Println("  (finalizer may have run)")
}

func main() {
	fmt.Println("=== Memory Pool, GC & Finalizers ===")
	syncPoolDemo()
	gcDemo()
	finalizerDemo()
}

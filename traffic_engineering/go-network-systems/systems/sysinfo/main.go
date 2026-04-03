// System information: CPU, memory, runtime stats.
package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	fmt.Println("=== System Information ===")

	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Arch: %s\n", runtime.GOARCH)
	fmt.Printf("  NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("  GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("  NumGoroutine: %d\n", runtime.NumGoroutine())
	fmt.Printf("  Go Version: %s\n", runtime.Version())

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("  Alloc: %d KB\n", memStats.Alloc/1024)
	fmt.Printf("  TotalAlloc: %d KB\n", memStats.TotalAlloc/1024)
	fmt.Printf("  Sys: %d KB\n", memStats.Sys/1024)
	fmt.Printf("  NumGC: %d\n", memStats.NumGC)
	fmt.Printf("  HeapObjects: %d\n", memStats.HeapObjects)

	cwd, _ := os.Getwd()
	fmt.Printf("  Working Dir: %s\n", cwd)

	hostname, _ := os.Hostname()
	fmt.Printf("  Hostname: %s\n", hostname)

	fmt.Printf("  TempDir: %s\n", os.TempDir())
}

// Stack vs Heap allocation and escape analysis in Go.
// Run with: go build -gcflags='-m' to see escape analysis output.
package main

import "fmt"

// This stays on the stack (no escape)
//
//go:noinline
func stackAllocation() int {
	x := 42
	return x
}

// This escapes to the heap (returned pointer)
//
//go:noinline
func heapAllocation() *int {
	x := 42
	return &x
}

// Slice that escapes
//
//go:noinline
func sliceEscape() []int {
	s := make([]int, 100)
	return s
}

// Slice that doesn't escape (small, stays on stack)
//
//go:noinline
func sliceNoEscape() int {
	s := make([]int, 3)
	s[0] = 1
	return s[0]
}

func main() {
	fmt.Println("=== Stack vs Heap Allocation ===")
	fmt.Println("  Run with: go build -gcflags='-m' to see escape analysis")

	_ = stackAllocation()
	p := heapAllocation()
	fmt.Printf("  Heap-allocated value: %d (at %p)\n", *p, p)

	s := sliceEscape()
	fmt.Printf("  Heap slice length: %d\n", len(s))

	v := sliceNoEscape()
	fmt.Printf("  Stack-local value: %d\n", v)
}

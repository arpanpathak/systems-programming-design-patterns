// Unsafe pointer operations, struct alignment, and cache line padding.
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("=== Unsafe Pointer Operations ===")

	// sizeof for common types
	fmt.Printf("  sizeof(int): %d\n", unsafe.Sizeof(int(0)))
	fmt.Printf("  sizeof(int64): %d\n", unsafe.Sizeof(int64(0)))
	fmt.Printf("  sizeof(string): %d\n", unsafe.Sizeof(""))
	fmt.Printf("  sizeof([]byte): %d\n", unsafe.Sizeof([]byte{}))

	// Struct layout and alignment
	type Example struct {
		a bool  // 1 byte
		b int64 // 8 bytes
		c bool  // 1 byte
		d int32 // 4 bytes
	}

	type ExampleOptimized struct {
		b int64 // 8 bytes
		d int32 // 4 bytes
		a bool  // 1 byte
		c bool  // 1 byte
	}

	fmt.Printf("\n--- Struct Padding ---\n")
	fmt.Printf("  Unoptimized struct size: %d bytes\n", unsafe.Sizeof(Example{}))
	fmt.Printf("  Optimized struct size:   %d bytes\n", unsafe.Sizeof(ExampleOptimized{}))

	// Field offsets
	fmt.Printf("  Example.a offset: %d\n", unsafe.Offsetof(Example{}.a))
	fmt.Printf("  Example.b offset: %d\n", unsafe.Offsetof(Example{}.b))
	fmt.Printf("  Example.c offset: %d\n", unsafe.Offsetof(Example{}.c))
	fmt.Printf("  Example.d offset: %d\n", unsafe.Offsetof(Example{}.d))

	// Pointer arithmetic
	arr := [5]int{10, 20, 30, 40, 50}
	p := unsafe.Pointer(&arr[0])
	elemSize := unsafe.Sizeof(arr[0])

	fmt.Printf("\n--- Pointer Arithmetic ---\n")
	for i := 0; i < 5; i++ {
		elemPtr := (*int)(unsafe.Pointer(uintptr(p) + uintptr(i)*elemSize))
		fmt.Printf("  arr[%d] = %d (via pointer arithmetic)\n", i, *elemPtr)
	}

	// Zero-copy string to []byte (DANGEROUS but fast)
	s := "hello world"
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	fmt.Printf("\n--- Zero-Copy ---\n")
	fmt.Printf("  Zero-copy string to bytes: %v\n", b)

	// Cache line / false sharing
	type PaddedCounter struct {
		value uint64
		_pad  [56]byte // Pad to 64 bytes (typical cache line size)
	}
	type UnpaddedCounters struct {
		a uint64
		b uint64
	}

	fmt.Printf("\n--- Cache Line / False Sharing ---\n")
	fmt.Printf("  Cache line size (typical): 64 bytes\n")
	fmt.Printf("  PaddedCounter size: %d bytes\n", unsafe.Sizeof(PaddedCounter{}))
	fmt.Printf("  UnpaddedCounters size: %d bytes (both in same cache line!)\n",
		unsafe.Sizeof(UnpaddedCounters{}))
	fmt.Println("  Padding prevents false sharing between goroutines")
}

package systems

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

// RunMMapExample demonstrates advanced OS File I/O using Memory-Mapped Files.
// standard `os.File` Read/Write requires copying data from kernel space to user space.
// mmap maps the file directly into the application's virtual memory address space.
// This is how high-performance databases (BoltDB, LMDB) and caching proxies
// achieve near-instant disk access.
func RunMMapExample() {
	fmt.Println("=== Advanced OS File I/O: Memory Mapping (mmap) ===")

	filename := "proxy_cache.db"

	// O_RDWR is required if we want to write to the mmap region.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Ensure the file has physical size on disk before mapping it!
	// We cannot map an empty file into memory sizes we don't own.
	size := int64(4096) // 1 standard Page Size
	if err := file.Truncate(size); err != nil {
		log.Fatalf("Failed to truncate (allocate space on disk): %v", err)
	}

	// Request memory map from the OS Kernel
	// PROT_READ | PROT_WRITE allows us to both read and mutually write to the array.
	// MAP_SHARED means changes propagate down to the actual physical disk file.
	b, err := unix.Mmap(int(file.Fd()), 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		log.Fatalf("Unix Mmap syscall failed: %v", err)
	}

	// Defer cleanup: System resources MUST be unmapped!
	defer func() {
		if err := unix.Munmap(b); err != nil {
			log.Printf("Unix Munmap failed: %v", err)
		}
	}()

	// Since the file is mapped, `b` is now essentially treating the Hard Drive as a byte array!
	msg := []byte("OS LEVEL KERNEL BYPASS I/O ACHIVED...")
	copy(b[0:], msg) // Write directly to RAM. The OS page synchronizer will flush it to disk later.

	// msync manually flushes the modified memory pages down to the physical disk (like fsync).
	if err := unix.Msync(b, unix.MS_SYNC); err != nil {
		log.Fatalf("msync syscall failed: %v", err)
	}

	fmt.Println("Successfully wrote to Memory-Mapped file using raw unix syscalls!")
	os.Remove(filename)
}

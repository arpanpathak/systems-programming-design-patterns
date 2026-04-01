package os_io

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// RunOSLevelIO demonstrates crucial Systems Programming concepts for I/O.
// Proxies (like Envoy) and databases (like etcd/Kubernetes) rely on these
// techniques for high-performance logging, WAL (Write-Ahead-Logs), and zero-copy transfers.
func RunOSLevelIO() {
	fmt.Println("=== OS-Level File I/O & Systems Programming ===")

	filename := "traffic_access.log"

	// 1. Raw OS File Descriptors & Flags
	// O_APPEND ensures atomic appends (multiple goroutines won't overwrite each other if OS supports it).
	// O_WRONLY avoids reading overhead, O_CREATE creates if missing.
	// 0644 are Unix permissions: Read/Write for owner, Read for others.
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("syscall failed to open file: %v\n", err)
	}
	// Always defer closing the File Descriptor to prevent FD leaks!
	defer file.Close()

	// 2. Buffered I/O (Mitigating Syscall Overhead)
	// Syscalls (context switching to kernel mode) are EXPENSIVE.
	// If a proxy logs every single request immediately, the CPU will choke.
	// bufio.Writer batches writes in user-space RAM until the buffer is full (or flushed).
	writer := bufio.NewWriterSize(file, 4096) // 4KB page-size buffer

	bytesWritten, err := writer.WriteString("Proxy Request 1: 200 OK\n")
	if err != nil {
		log.Printf("write failed: %v\n", err)
	}
	fmt.Printf("Buffered %d bytes in user-space...\n", bytesWritten)

	// Since it's buffered, we MUST flush it to push the bytes into the kernel's OS buffer cache.
	writer.Flush()
	fmt.Println("Flushed user-space buffer to Kernel Page Cache.")

	// 3. Fsync (Guaranteeing Persistence)
	// Even after flushing, the OS delays writing to the physical SSD/HDD.
	// If the server loses power now, data is lost!
	// file.Sync() executes the `fsync` syscall, forcing the hardware disk to write immediately.
	// Distributed systems (like Raft or Paxos) require fsync for their durability guarantees.
	err = file.Sync()
	if err != nil {
		log.Fatalf("fsync syscall failed: %v\n", err)
	}
	fmt.Println("Executed fsync syscall. Data is safely on the physical disk.")

	// 4. Zero-Copy Concept (io.Copy)
	// When proxies serve static files or stream data from disk to network,
	// reading to user-space and writing to socket is slow (2 context switches, 2 CPU copies).
	// io.Copy in Go often uses the `sendfile` syscall under the hood on Linux,
	// bypassing user-space entirely! (Zero-Copy)
	runZeroCopyTransfer(filename)
}

func runZeroCopyTransfer(srcFilename string) {
	srcBuf, err := os.Open(srcFilename)
	if err != nil {
		return
	}
	defer srcBuf.Close()

	// Create a dummy destination (e.g. simulating a TCP socket in a proxy)
	dstFilename := "traffic_access_copy.log"
	dstBuf, err := os.Create(dstFilename)
	if err != nil {
		return
	}
	defer dstBuf.Close()

	// In Go, io.Copy between os.File and a TCP socket will utilize `sendfile` automatically
	// on supported platforms, giving maximum proxying performance.
	copied, err := io.Copy(dstBuf, srcBuf)
	if err != nil {
		log.Printf("io.Copy failed: %v\n", err)
	} else {
		fmt.Printf("Zero-copy transferred %d bytes via io.Copy optimization.\n", copied)
	}

	// Cleanup testing files
	os.Remove("traffic_access.log")
	os.Remove("traffic_access_copy.log")
}

//go:build darwin

package systems

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

// RunRawSocketExample demonstrates bypassing the Go 'net' abstraction
// and speaking directly to the Darwin OS Kernel via Syscalls.
func RunRawSocketExample() {
	fmt.Println("=== OS Level Socket Provisioning (Darwin) ===")

	// 1. Create a socket file descriptor (FD) just like C/C++.
	// AF_INET = IPv4, SOCK_STREAM = TCP.
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalf("unix.Socket failed: %v", err)
	}
	// Socket FDs must be closed to prevent FD Exhaustion (ulimit violations)
	defer unix.Close(fd)

	// 2. SO_REUSEADDR allows restarting the proxy instantly without "Address already in use" errors.
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		log.Fatalf("Setsockopt SO_REUSEADDR failed: %v", err)
	}

	// Set to non-blocking mode (CRITICAL for Epoll/Kqueue event loops)
	if err := unix.SetNonblock(fd, true); err != nil {
		log.Fatalf("SetNonblock failed: %v", err)
	}

	// 3. Bind the socket to Port 8080.
	sockAddr := &unix.SockaddrInet4{
		Port: 8080,
		Addr: [4]byte{0, 0, 0, 0}, // 0.0.0.0 (All interfaces)
	}

	if err := unix.Bind(fd, sockAddr); err != nil {
		log.Fatalf("unix.Bind failed: %v", err)
	}

	// 4. Listen instructs the Kernel to accept incoming SYN packets.
	// The backlog of 1024 represents max un-accepted connection queue before Kernel drops SYNs.
	if err := unix.Listen(fd, 1024); err != nil {
		log.Fatalf("unix.Listen failed: %v", err)
	}

	fmt.Println("Successfully created raw, non-blocking Kernel Socket waiting on Port 8080!")
	fmt.Println("Ready to be integrated with a Kqueue/Epoll event loop.")
}

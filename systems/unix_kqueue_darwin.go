//go:build darwin

package systems

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

// RunKqueueEventLoop demonstrates Mac/BSD's equivalent to Linux 'epoll'.
// This is the absolute core of systems programming for networking.
// Instead of 10,000 goroutines waiting on 10,000 idle TCP connections (consuming memory),
// a proxy uses Kqueue/Epoll to utilize ONE thread that gets notified by the Kernel
// ONLY when a socket actually receives bytes!
func RunKqueueEventLoop(serverFD int) {
	fmt.Println("=== Initializing Darwin Kqueue (Event Loop) ===")

	// 1. Create a Kernel Queue Descriptor
	kqFD, err := unix.Kqueue()
	if err != nil {
		log.Fatalf("Failed to create kqueue: %v\n", err)
	}
	defer unix.Close(kqFD)

	// 2. Register an Event Filter for our server socket
	// EVFILT_READ means we want the kernel to wake us up when there is data to read.
	// EV_ADD adds it. EV_ENABLE turns the filter on.
	changeEvent := unix.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: unix.EVFILT_READ,
		Flags:  unix.EV_ADD | unix.EV_ENABLE,
	}

	// Submit the change request to Kqueue
	_, err = unix.Kevent(kqFD, []unix.Kevent_t{changeEvent}, nil, nil)
	if err != nil {
		log.Fatalf("Failed to register kqueue event monitor: %v\n", err)
	}

	fmt.Println("Kqueue initialized. Starting highly scalable I/O Event Loop...")

	// 3. The Grand OS Event Loop
	events := make([]unix.Kevent_t, 32) // Array to read fired kernel events into

	for {
		// This syscall BLOCKS the thread until the OS Network Stack receives data!
		// 'n' is the number of file descriptors that have action waiting.
		n, err := unix.Kevent(kqFD, nil, events, nil)
		if err != nil {
			if err == unix.EINTR {
				continue // Interrupted by signal, ignore
			}
			log.Fatalf("kqueue wait failed: %v", err)
		}

		for i := 0; i < n; i++ {
			event := events[i]
			fd := int(event.Ident)

			if fd == serverFD {
				// The Server Socket woke up -> A client executed a TCP Handshake!
				nfd, _, err := unix.Accept(serverFD)
				if err != nil {
					log.Printf("unix.Accept error: %v", err)
					continue
				}

				fmt.Printf("Accepted client connection (FD: %d) through Kernel Event Loop\n", nfd)

				unix.SetNonblock(nfd, true)
				// Here, we would insert `nfd` BACK into Kqueue to monitor for downstream client reads
				// in a real C/C++ proxy codebase.
				unix.Close(nfd)
			}
		}

		// Break out of infinite loop for this example demo
		break
	}
}

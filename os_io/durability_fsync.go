package os_io

import (
	"fmt"
	"log"
	"os"
)

// RunFsyncDurability demonstrates the absolute gold-standard of File I/O for
// systems that CANNOT lose data, even if the OS panics or the rack loses power.
// Envoy proxies use this for dumping access logs and Control Planes use it for WALs.
func RunFsyncDurability() {
	fmt.Println("=== OS-Level Durability: Fsync & Atomic Appends ===")

	filename := "write_ahead_log.bin"

	// 1. O_APPEND ensures atomic appends (multiple goroutines won't overwrite each other
	// if the OS POSIX standard supports it).
	// O_WRONLY avoids reading overhead, O_CREATE creates if missing.
	// O_SYNC (Alternative to fsync call): Opens the file with synchronous I/O enabled natively.
	// But O_SYNC on every open is painfully slow, so we prefer manual file.Sync().
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("syscall failed to open file: %v\n", err)
	}
	defer file.Close()

	// 2. Write Data to the Kernel Buffer Cache
	bytesWritten, err := file.WriteString("Proxy Traffic Commit Event: 0x1A4F\n")
	if err != nil {
		log.Printf("write failed: %v\n", err)
	}
	fmt.Printf("Written %d bytes to Kernel Page Cache.\n", bytesWritten)

	// 3. Guaranteeing Persistence
	// Even after writing, the OS delays writing to the physical SSD/HDD to boost performance.
	// If the server loses power NOW, data is lost permanently!
	// file.Sync() executes the `fsync` syscall, forcing the hardware disk to write immediately
	// and spin indefinitely until the magnetic strip/SSD controller confirms it is saved.
	err = file.Sync()
	if err != nil {
		log.Fatalf("fsync syscall failed: %v\n", err)
	}

	fmt.Println("Executed fsync syscall successfully. Data is safely bound to physical disk geometry.")

	// Clean up demo
	os.Remove(filename)
}

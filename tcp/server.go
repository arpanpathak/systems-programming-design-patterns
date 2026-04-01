package tcp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// RunTCPServer is a core systems programming example demonstrating how edge servers
// like Envoy or NGINX accept raw byte streams, frame them, and handle timeouts.
func RunTCPServer(address string) {
	// Listen on TCP IPv4/IPv6 port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to bind TCP port: %v\n", err)
	}
	defer listener.Close()

	log.Printf("TCP Server actively listening on %s...\n", address)

	for {
		// Accept incoming connection (blocking call)
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		// Handle each connection concurrently in its own goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New TCP connection established from %s\n", remoteAddr)

	defer func() {
		log.Printf("Closing TCP connection for %s\n", remoteAddr)
		_ = conn.Close()
	}()

	// Apply Keep-Alive and Read deadlines to prevent "Slowloris" attacks!
	// Real-world proxies strictly configure connection drop horizons.
	deadlineDuration := 30 * time.Second
	_ = conn.SetReadDeadline(time.Now().Add(deadlineDuration))

	// Buffer reader for handling arbitrary packet framing
	reader := bufio.NewReader(conn)

	for {
		// Read until newline (simplistic framing)
		message, err := reader.ReadString('\n')
		if err != nil {
			// Expected when connection is closed remotely cleanly (EOF).
			if err != io.EOF {
				log.Printf("Read error for %s: %v\n", remoteAddr, err)
			}
			break
		}

		// Refresh the deadline on successful packet read
		_ = conn.SetReadDeadline(time.Now().Add(deadlineDuration))

		log.Printf("Received stream from %s: %s", remoteAddr, message)

		// Acknowledge receipt
		ack := []byte(fmt.Sprintf("ACK: %s", message))
		_, writeErr := conn.Write(ack)
		if writeErr != nil {
			log.Printf("Write failed for %s: %v\n", remoteAddr, writeErr)
			break
		}
	}
}

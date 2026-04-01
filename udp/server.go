package udp

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// sync.Pool reduces Garbage Collection (GC) pressure by recycling memory buffers.
// In high-throughput, latency-critical systems (like DNS or QUIC proxies),
// frequently allocating buffers is a lethal bottleneck.
var bufferPool = sync.Pool{
	New: func() any {
		// 2048 bytes accommodates most standardized MTUs
		buf := make([]byte, 2048)
		return &buf
	},
}

// RunUDPServer listens for datagrams on a specific port and responds statelessly.
func RunUDPServer(address string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v\n", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to bind UDP socket: %v\n", err)
	}
	defer conn.Close()

	log.Printf("UDP Server actively listening on %s...\n", address)

	// Since UDP is connectionless, one goroutine reads from the socket,
	// but we can spawn worker goroutines to handle the payload independently!
	for {
		// Acquire a pre-allocated buffer from our Pool
		poolBuf := bufferPool.Get().(*[]byte)
		buffer := *poolBuf

		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("ReadFromUDP error: %v\n", err)
			bufferPool.Put(poolBuf) // Recycle buffer on failure
			continue
		}

		// Dispatch datagram to worker (Multiplexing)
		go handleDatagram(conn, remoteAddr, poolBuf, n)
	}
}

func handleDatagram(conn *net.UDPConn, remoteAddr *net.UDPAddr, poolBuf *[]byte, n int) {
	// Crucial: We must recycle the buffer back to the pool AFTER processing finishes
	defer bufferPool.Put(poolBuf)

	payload := (*poolBuf)[:n] // Slice up to read bytes

	log.Printf("Received %d bytes datagram from %s\n", n, remoteAddr.String())

	// E.g. proxying the UDP packet logic here...

	// Acknowledge Datagram
	response := []byte(fmt.Sprintf("UDP-ACK: %s", string(payload)))
	_, err := conn.WriteToUDP(response, remoteAddr)
	if err != nil {
		log.Printf("WriteToUDP error answering %s: %v\n", remoteAddr.String(), err)
	}
}

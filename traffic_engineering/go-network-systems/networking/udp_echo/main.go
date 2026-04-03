// UDP Echo Server/Client with datagram basics.
package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type UDPEchoServer struct {
	conn *net.UDPConn
	quit chan struct{}
	wg   sync.WaitGroup
}

func NewUDPEchoServer(addr string) (*UDPEchoServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPEchoServer{conn: conn, quit: make(chan struct{})}, nil
}

func (s *UDPEchoServer) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		buf := make([]byte, 65535)
		for {
			select {
			case <-s.quit:
				return
			default:
			}
			s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, remoteAddr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}
			s.conn.WriteToUDP(buf[:n], remoteAddr)
		}
	}()
}

func (s *UDPEchoServer) Shutdown() {
	close(s.quit)
	s.conn.Close()
	s.wg.Wait()
}

func main() {
	fmt.Println("=== UDP Echo Server/Client ===")

	server, err := NewUDPEchoServer("127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Server error: %v\n", err)
		return
	}
	server.Start()
	defer server.Shutdown()

	serverAddr := server.conn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("  UDP server on %s\n", serverAddr)

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("  Dial error: %v\n", err)
		return
	}
	defer conn.Close()

	for _, msg := range []string{"hello", "udp", "packets"} {
		conn.Write([]byte(msg))
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("  Read error: %v\n", err)
			continue
		}
		fmt.Printf("  Sent: %q -> Echo: %q\n", msg, string(buf[:n]))
	}

	// MTU reference
	fmt.Println("\n--- MTU Reference ---")
	fmt.Println("  Ethernet MTU: 1500 bytes")
	fmt.Println("  Max UDP payload (IPv4): 1500 - 20 - 8 = 1472 bytes")
	fmt.Println("  Max UDP payload (IPv6): 1500 - 40 - 8 = 1452 bytes")
	fmt.Println("  Absolute max UDP: 65535 - 20 - 8 = 65507 bytes")
}

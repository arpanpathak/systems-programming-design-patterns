// TCP Echo Server with graceful shutdown, keep-alive, and deadline handling.
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPServer struct {
	listener net.Listener
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewTCPServer(addr string) (*TCPServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPServer{listener: l, quit: make(chan struct{})}, nil
}

func (s *TCPServer) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					fmt.Printf("  Accept error: %v\n", err)
					continue
				}
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn)
			}()
		}
	}()
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// TCP keep-alive and options
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true) // Disable Nagle's (TCP_NODELAY)
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "QUIT" {
			fmt.Fprintf(conn, "BYE\n")
			return
		}
		conn.Write([]byte(strings.ToUpper(line) + "\n"))
		conn.SetDeadline(time.Now().Add(30 * time.Second))
	}
}

func (s *TCPServer) Shutdown() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

type TCPClient struct {
	addr       string
	conn       net.Conn
	mu         sync.Mutex
	maxRetries int
}

func NewTCPClient(addr string) *TCPClient {
	return &TCPClient{addr: addr, maxRetries: 3}
}

func (c *TCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	for i := 0; i < c.maxRetries; i++ {
		c.conn, err = net.DialTimeout("tcp", c.addr, 5*time.Second)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(1<<uint(i)) * 100 * time.Millisecond)
	}
	return fmt.Errorf("failed after %d retries: %w", c.maxRetries, err)
}

func (c *TCPClient) Send(msg string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	fmt.Fprintf(c.conn, "%s\n", msg)
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}

func (c *TCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func main() {
	fmt.Println("=== TCP Echo Server/Client ===")

	server, err := NewTCPServer("127.0.0.1:0")
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
		return
	}
	server.Start()
	defer server.Shutdown()
	fmt.Printf("  Echo server on %s\n", server.listener.Addr())

	client := NewTCPClient(server.listener.Addr().String())
	if err := client.Connect(); err != nil {
		fmt.Printf("  Connect error: %v\n", err)
		return
	}
	defer client.Close()

	for _, msg := range []string{"hello", "world", "network traffic"} {
		resp, err := client.Send(msg)
		if err != nil {
			fmt.Printf("  Send error: %v\n", err)
			break
		}
		fmt.Printf("  Sent: %q -> Received: %q\n", msg, resp)
	}
}

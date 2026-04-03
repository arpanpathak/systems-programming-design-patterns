// TCP Layer 4 proxy — bidirectional byte relay.
package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type TCPProxy struct {
	listenAddr string
	targetAddr string
	listener   net.Listener
	quit       chan struct{}
	wg         sync.WaitGroup
}

func NewTCPProxy(listenAddr, targetAddr string) *TCPProxy {
	return &TCPProxy{listenAddr: listenAddr, targetAddr: targetAddr, quit: make(chan struct{})}
}

func (p *TCPProxy) Start() error {
	var err error
	p.listener, err = net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			conn, err := p.listener.Accept()
			if err != nil {
				select {
				case <-p.quit:
					return
				default:
					continue
				}
			}
			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				p.relay(conn)
			}()
		}
	}()
	return nil
}

func (p *TCPProxy) relay(clientConn net.Conn) {
	defer clientConn.Close()
	targetConn, err := net.DialTimeout("tcp", p.targetAddr, 10*time.Second)
	if err != nil {
		fmt.Printf("  TCP Proxy: dial error: %v\n", err)
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(targetConn, clientConn)
		if tc, ok := targetConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, targetConn)
		if tc, ok := clientConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()
	wg.Wait()
}

func (p *TCPProxy) Shutdown() {
	close(p.quit)
	p.listener.Close()
	p.wg.Wait()
}

func main() {
	fmt.Println("=== TCP Proxy (Layer 4) ===")

	// Start a backend
	backend, _ := net.Listen("tcp", "127.0.0.1:0")
	defer backend.Close()
	go func() {
		for {
			conn, err := backend.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				buf := make([]byte, 1024)
				n, _ := conn.Read(buf)
				conn.Write([]byte("BACKEND:" + string(buf[:n])))
			}()
		}
	}()

	proxy := NewTCPProxy("127.0.0.1:0", backend.Addr().String())
	if err := proxy.Start(); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer proxy.Shutdown()

	fmt.Printf("  Proxy %s -> %s\n", proxy.listener.Addr(), backend.Addr())

	conn, _ := net.Dial("tcp", proxy.listener.Addr().String())
	defer conn.Close()
	conn.Write([]byte("hello"))
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := conn.Read(buf)
	fmt.Printf("  Through proxy: %q\n", string(buf[:n]))
}

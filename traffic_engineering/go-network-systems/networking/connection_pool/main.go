// TCP Connection Pool with idle-connection reuse and health checks.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type ConnPool struct {
	mu      sync.Mutex
	conns   chan net.Conn
	factory func() (net.Conn, error)
	maxSize int
	active  int
}

func NewConnPool(maxSize int, factory func() (net.Conn, error)) *ConnPool {
	return &ConnPool{
		conns:   make(chan net.Conn, maxSize),
		factory: factory,
		maxSize: maxSize,
	}
}

func (p *ConnPool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case conn := <-p.conns:
		conn.SetReadDeadline(time.Now())
		one := make([]byte, 1)
		if _, err := conn.Read(one); err == io.EOF {
			conn.Close()
			return p.factory()
		}
		conn.SetReadDeadline(time.Time{})
		return conn, nil
	default:
	}

	p.mu.Lock()
	if p.active < p.maxSize {
		p.active++
		p.mu.Unlock()
		conn, err := p.factory()
		if err != nil {
			p.mu.Lock()
			p.active--
			p.mu.Unlock()
			return nil, err
		}
		return conn, nil
	}
	p.mu.Unlock()

	select {
	case conn := <-p.conns:
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *ConnPool) Put(conn net.Conn) {
	select {
	case p.conns <- conn:
	default:
		conn.Close()
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
	}
}

func (p *ConnPool) Close() {
	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}
}

// Simple echo server for testing
func startEchoServer() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				scanner := bufio.NewScanner(conn)
				for scanner.Scan() {
					conn.Write([]byte(strings.ToUpper(scanner.Text()) + "\n"))
				}
			}()
		}
	}()
	return l, nil
}

func main() {
	fmt.Println("=== Connection Pool ===")

	l, err := startEchoServer()
	if err != nil {
		fmt.Printf("  Server error: %v\n", err)
		return
	}
	defer l.Close()

	addr := l.Addr().String()
	pool := NewConnPool(5, func() (net.Conn, error) {
		return net.DialTimeout("tcp", addr, 5*time.Second)
	})
	defer pool.Close()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, err := pool.Get(ctx)
			if err != nil {
				fmt.Printf("  Worker %d: pool error: %v\n", id, err)
				return
			}

			msg := fmt.Sprintf("hello from worker %d", id)
			fmt.Fprintf(conn, "%s\n", msg)

			reader := bufio.NewReader(conn)
			resp, _ := reader.ReadString('\n')
			fmt.Printf("  Worker %d: %s -> %s", id, msg, resp)

			pool.Put(conn)
		}(i)
	}
	wg.Wait()
}

// Line-based Key-Value protocol over TCP.
// Commands: SET key value, GET key, DEL key, QUIT
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type KVServer struct {
	listener net.Listener
	store    sync.Map
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewKVServer(addr string) (*KVServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &KVServer{listener: l, quit: make(chan struct{})}, nil
}

func (s *KVServer) Start() {
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
					continue
				}
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConn(conn)
			}()
		}
	}()
}

func (s *KVServer) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		switch strings.ToUpper(parts[0]) {
		case "SET":
			if len(parts) >= 3 {
				s.store.Store(parts[1], strings.Join(parts[2:], " "))
				fmt.Fprintf(conn, "OK\n")
			} else {
				fmt.Fprintf(conn, "ERR: SET key value\n")
			}
		case "GET":
			if len(parts) >= 2 {
				if val, ok := s.store.Load(parts[1]); ok {
					fmt.Fprintf(conn, "%v\n", val)
				} else {
					fmt.Fprintf(conn, "NOT_FOUND\n")
				}
			}
		case "DEL":
			if len(parts) >= 2 {
				s.store.Delete(parts[1])
				fmt.Fprintf(conn, "OK\n")
			}
		case "QUIT":
			fmt.Fprintf(conn, "BYE\n")
			return
		default:
			fmt.Fprintf(conn, "ERR: unknown command\n")
		}
	}
}

func (s *KVServer) Shutdown() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func main() {
	fmt.Println("=== Key-Value Protocol Server ===")

	server, err := NewKVServer("127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	server.Start()
	defer server.Shutdown()
	fmt.Printf("  KV server on %s\n", server.listener.Addr())

	conn, err := net.Dial("tcp", server.listener.Addr().String())
	if err != nil {
		fmt.Printf("  Dial error: %v\n", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	commands := []string{
		"SET name Go",
		"SET version 1.21",
		"GET name",
		"GET version",
		"GET nonexistent",
		"DEL name",
		"GET name",
		"QUIT",
	}
	for _, cmd := range commands {
		fmt.Fprintf(conn, "%s\n", cmd)
		resp, _ := reader.ReadString('\n')
		fmt.Printf("  > %s -> %s", cmd, resp)
	}
}

// Event-driven server (reactor pattern) with single event loop + stats.
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type EventType int

const (
	EventAccept EventType = iota
	EventRead
	EventClose
)

type ConnEvent struct {
	Type EventType
	Conn net.Conn
	Data []byte
}

type EventDrivenServer struct {
	listener    net.Listener
	events      chan ConnEvent
	connections sync.Map
	quit        chan struct{}
	wg          sync.WaitGroup
	handler     func([]byte) []byte

	TotalConns   int64
	ActiveConns  int64
	BytesIn      int64
	BytesOut     int64
	RequestCount int64
}

func NewEventDrivenServer(addr string, handler func([]byte) []byte) (*EventDrivenServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &EventDrivenServer{
		listener: l,
		events:   make(chan ConnEvent, 1024),
		quit:     make(chan struct{}),
		handler:  handler,
	}, nil
}

func (s *EventDrivenServer) Start() {
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
			s.events <- ConnEvent{Type: EventAccept, Conn: conn}
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.quit:
				return
			case event := <-s.events:
				s.processEvent(event)
			}
		}
	}()
}

func (s *EventDrivenServer) processEvent(event ConnEvent) {
	switch event.Type {
	case EventAccept:
		atomic.AddInt64(&s.TotalConns, 1)
		atomic.AddInt64(&s.ActiveConns, 1)
		s.connections.Store(event.Conn, true)
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			scanner := bufio.NewScanner(event.Conn)
			for scanner.Scan() {
				data := make([]byte, len(scanner.Bytes()))
				copy(data, scanner.Bytes())
				atomic.AddInt64(&s.BytesIn, int64(len(data)))
				s.events <- ConnEvent{Type: EventRead, Conn: event.Conn, Data: data}
			}
			s.events <- ConnEvent{Type: EventClose, Conn: event.Conn}
		}()

	case EventRead:
		atomic.AddInt64(&s.RequestCount, 1)
		response := s.handler(event.Data)
		n, err := event.Conn.Write(response)
		if err != nil {
			s.events <- ConnEvent{Type: EventClose, Conn: event.Conn}
			return
		}
		atomic.AddInt64(&s.BytesOut, int64(n))

	case EventClose:
		if _, loaded := s.connections.LoadAndDelete(event.Conn); loaded {
			atomic.AddInt64(&s.ActiveConns, -1)
			event.Conn.Close()
		}
	}
}

func (s *EventDrivenServer) Shutdown() {
	close(s.quit)
	s.listener.Close()
	s.connections.Range(func(key, _ interface{}) bool {
		key.(net.Conn).Close()
		return true
	})
	s.wg.Wait()
}

func main() {
	fmt.Println("=== Event-Driven Server (Reactor Pattern) ===")

	server, err := NewEventDrivenServer("127.0.0.1:0", func(data []byte) []byte {
		return []byte(strings.ToUpper(string(data)) + "\n")
	})
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	server.Start()
	defer server.Shutdown()

	addr := server.listener.Addr().String()
	fmt.Printf("  Server on %s\n", addr)

	var wg sync.WaitGroup
	for c := 0; c < 5; c++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			defer conn.Close()
			reader := bufio.NewReader(conn)
			for i := 0; i < 3; i++ {
				msg := fmt.Sprintf("client%d-msg%d", id, i)
				fmt.Fprintf(conn, "%s\n", msg)
				resp, _ := reader.ReadString('\n')
				fmt.Printf("  Client %d: %s -> %s", id, msg, resp)
			}
		}(c)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("  Stats: conns=%d active=%d reqs=%d in=%d out=%d\n",
		atomic.LoadInt64(&server.TotalConns),
		atomic.LoadInt64(&server.ActiveConns),
		atomic.LoadInt64(&server.RequestCount),
		atomic.LoadInt64(&server.BytesIn),
		atomic.LoadInt64(&server.BytesOut))

	fmt.Println(`
  Go's netpoller uses OS-specific multiplexing:
  - Linux: epoll | macOS: kqueue | Windows: IOCP
  
  Reactor: "I'll tell you when FD is ready" (epoll/kqueue)
  Proactor: "I'll do the I/O, tell you when done" (IOCP)

  Key socket options:
  SO_REUSEADDR, SO_REUSEPORT, TCP_NODELAY, SO_KEEPALIVE,
  TCP_FASTOPEN, SO_LINGER, TCP_CORK, TCP_QUICKACK`)
}

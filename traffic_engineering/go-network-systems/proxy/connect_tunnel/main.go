// HTTP CONNECT tunnel — for HTTPS proxying.
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type ConnectProxy struct {
	listener net.Listener
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewConnectProxy(addr string) (*ConnectProxy, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &ConnectProxy{listener: l, quit: make(chan struct{})}, nil
}

func (p *ConnectProxy) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		server := &http.Server{Handler: http.HandlerFunc(p.handleConnect)}
		go server.Serve(p.listener)
		<-p.quit
		server.Close()
	}()
}

func (p *ConnectProxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Printf("  CONNECT tunnel to: %s\n", r.Host)

	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		targetConn.Close()
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		targetConn.Close()
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(targetConn, clientConn); targetConn.Close() }()
	go func() { defer wg.Done(); io.Copy(clientConn, targetConn); clientConn.Close() }()
	wg.Wait()
}

func (p *ConnectProxy) Shutdown() {
	close(p.quit)
	p.listener.Close()
	p.wg.Wait()
}

func main() {
	fmt.Println("=== HTTP CONNECT Proxy ===")

	proxy, err := NewConnectProxy("127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	proxy.Start()
	defer proxy.Shutdown()

	fmt.Printf("  CONNECT proxy on %s\n", proxy.listener.Addr())

	fmt.Println(`
  How CONNECT tunneling works:
  1. Client sends: CONNECT host:443 HTTP/1.1
  2. Proxy connects to target on port 443
  3. Proxy responds: HTTP/1.1 200 Connection Established
  4. Proxy relays raw bytes bidirectionally
  5. TLS handshake happens through the tunnel
  Proxy CANNOT see encrypted traffic (end-to-end TLS)`)
}

// Full SOCKS5 proxy server (RFC 1928) with optional username/password auth.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	SOCKS5Version         = 0x05
	AuthNone              = 0x00
	AuthPassword          = 0x02
	AuthNoAccept          = 0xFF
	CmdConnect            = 0x01
	CmdBind               = 0x02
	CmdUDP                = 0x03
	AddrIPv4              = 0x01
	AddrDomain            = 0x03
	AddrIPv6              = 0x04
	ReplySuccess          = 0x00
	ReplyHostUnreach      = 0x04
	ReplyCmdNotSupported  = 0x07
	ReplyAddrNotSupported = 0x08
)

type SOCKS5Server struct {
	listener    net.Listener
	quit        chan struct{}
	wg          sync.WaitGroup
	credentials map[string]string
}

func NewSOCKS5Server(addr string, creds map[string]string) (*SOCKS5Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &SOCKS5Server{listener: l, quit: make(chan struct{}), credentials: creds}, nil
}

func (s *SOCKS5Server) Start() {
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
				s.handleClient(conn)
			}()
		}
	}()
}

func (s *SOCKS5Server) handleClient(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	if err := s.negotiate(conn); err != nil {
		return
	}
	s.handleRequest(conn)
}

func (s *SOCKS5Server) negotiate(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != SOCKS5Version {
		return fmt.Errorf("unsupported version: %d", header[0])
	}
	methods := make([]byte, header[1])
	io.ReadFull(conn, methods)

	if s.credentials != nil {
		if !containsByte(methods, AuthPassword) {
			conn.Write([]byte{SOCKS5Version, AuthNoAccept})
			return fmt.Errorf("no acceptable auth")
		}
		conn.Write([]byte{SOCKS5Version, AuthPassword})
		return s.authenticatePassword(conn)
	}
	if !containsByte(methods, AuthNone) {
		conn.Write([]byte{SOCKS5Version, AuthNoAccept})
		return fmt.Errorf("no acceptable auth")
	}
	conn.Write([]byte{SOCKS5Version, AuthNone})
	return nil
}

func (s *SOCKS5Server) authenticatePassword(conn net.Conn) error {
	header := make([]byte, 2)
	io.ReadFull(conn, header)
	username := make([]byte, header[1])
	io.ReadFull(conn, username)
	plen := make([]byte, 1)
	io.ReadFull(conn, plen)
	password := make([]byte, plen[0])
	io.ReadFull(conn, password)

	if pass, ok := s.credentials[string(username)]; ok && pass == string(password) {
		conn.Write([]byte{0x01, 0x00})
		return nil
	}
	conn.Write([]byte{0x01, 0x01})
	return fmt.Errorf("auth failed")
}

func (s *SOCKS5Server) handleRequest(conn net.Conn) error {
	header := make([]byte, 4)
	io.ReadFull(conn, header)
	if header[0] != SOCKS5Version {
		return fmt.Errorf("invalid version")
	}

	var destAddr string
	switch header[3] {
	case AddrIPv4:
		addr := make([]byte, 4)
		io.ReadFull(conn, addr)
		destAddr = net.IP(addr).String()
	case AddrDomain:
		lenBuf := make([]byte, 1)
		io.ReadFull(conn, lenBuf)
		domain := make([]byte, lenBuf[0])
		io.ReadFull(conn, domain)
		destAddr = string(domain)
	case AddrIPv6:
		addr := make([]byte, 16)
		io.ReadFull(conn, addr)
		destAddr = net.IP(addr).String()
	default:
		s.sendReply(conn, ReplyAddrNotSupported, nil)
		return fmt.Errorf("unsupported addr type")
	}

	portBuf := make([]byte, 2)
	io.ReadFull(conn, portBuf)
	port := binary.BigEndian.Uint16(portBuf)
	target := fmt.Sprintf("%s:%d", destAddr, port)
	fmt.Printf("  SOCKS5: CONNECT -> %s\n", target)

	if header[1] != CmdConnect {
		s.sendReply(conn, ReplyCmdNotSupported, nil)
		return fmt.Errorf("unsupported cmd")
	}

	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		s.sendReply(conn, ReplyHostUnreach, nil)
		return err
	}
	defer targetConn.Close()

	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	s.sendReply(conn, ReplySuccess, localAddr)
	conn.SetDeadline(time.Time{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(targetConn, conn) }()
	go func() { defer wg.Done(); io.Copy(conn, targetConn) }()
	wg.Wait()
	return nil
}

func (s *SOCKS5Server) sendReply(conn net.Conn, reply byte, addr *net.TCPAddr) {
	resp := []byte{SOCKS5Version, reply, 0x00}
	if addr != nil && addr.IP.To4() != nil {
		resp = append(resp, AddrIPv4)
		resp = append(resp, addr.IP.To4()...)
	} else {
		resp = append(resp, AddrIPv4, 0, 0, 0, 0)
	}
	portBuf := make([]byte, 2)
	if addr != nil {
		binary.BigEndian.PutUint16(portBuf, uint16(addr.Port))
	}
	resp = append(resp, portBuf...)
	conn.Write(resp)
}

func (s *SOCKS5Server) Shutdown() { close(s.quit); s.listener.Close(); s.wg.Wait() }

func containsByte(s []byte, b byte) bool {
	for _, v := range s {
		if v == b {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("=== SOCKS5 Proxy Server ===")

	// Backend echo server
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
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					conn.Write([]byte("ECHO:" + string(buf[:n])))
				}
			}()
		}
	}()

	socks, _ := NewSOCKS5Server("127.0.0.1:0", nil)
	socks.Start()
	defer socks.Shutdown()
	fmt.Printf("  SOCKS5 server on %s\n", socks.listener.Addr())

	// Manual SOCKS5 client handshake
	conn, _ := net.Dial("tcp", socks.listener.Addr().String())
	defer conn.Close()

	conn.Write([]byte{0x05, 0x01, 0x00}) // Greeting
	resp := make([]byte, 2)
	io.ReadFull(conn, resp)
	fmt.Printf("  Auth: version=%d method=%d\n", resp[0], resp[1])

	backendAddr := backend.Addr().(*net.TCPAddr)
	req := []byte{0x05, 0x01, 0x00, AddrIPv4}
	req = append(req, backendAddr.IP.To4()...)
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, uint16(backendAddr.Port))
	req = append(req, portBuf...)
	conn.Write(req)

	reply := make([]byte, 10)
	io.ReadFull(conn, reply)
	fmt.Printf("  Reply: version=%d status=%d\n", reply[0], reply[1])

	if reply[1] == ReplySuccess {
		conn.Write([]byte("hello via socks5"))
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if err == nil {
			fmt.Printf("  Through SOCKS5: %q\n", string(buf[:n]))
		}
	}

	fmt.Println(`
  SOCKS5 (RFC 1928):
  - Layer 5 (session), supports TCP + UDP
  - Auth: None, Username/Password (RFC 1929)
  - Address types: IPv4, IPv6, Domain
  - Commands: CONNECT, BIND, UDP ASSOCIATE
  vs HTTP CONNECT: SOCKS5 is binary/minimal, works at lower layer`)
}

// Reliable UDP using Stop-and-Wait ARQ protocol.
package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type ReliablePacket struct {
	SeqNum  uint32
	IsACK   bool
	Payload []byte
}

func (p *ReliablePacket) Marshal() []byte {
	buf := make([]byte, 5+len(p.Payload))
	binary.BigEndian.PutUint32(buf[0:4], p.SeqNum)
	if p.IsACK {
		buf[4] = 1
	}
	copy(buf[5:], p.Payload)
	return buf
}

func UnmarshalPacket(data []byte) (*ReliablePacket, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("packet too short")
	}
	return &ReliablePacket{
		SeqNum:  binary.BigEndian.Uint32(data[0:4]),
		IsACK:   data[4] == 1,
		Payload: data[5:],
	}, nil
}

type ReliableUDPSender struct {
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	seqNum     uint32
	timeout    time.Duration
	maxRetries int
}

func NewReliableUDPSender(conn *net.UDPConn, remote *net.UDPAddr) *ReliableUDPSender {
	return &ReliableUDPSender{
		conn:       conn,
		remoteAddr: remote,
		timeout:    200 * time.Millisecond,
		maxRetries: 5,
	}
}

func (s *ReliableUDPSender) Send(payload []byte) error {
	pkt := &ReliablePacket{SeqNum: s.seqNum, Payload: payload}
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		s.conn.WriteToUDP(pkt.Marshal(), s.remoteAddr)
		s.conn.SetReadDeadline(time.Now().Add(s.timeout))
		buf := make([]byte, 1024)
		n, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("    Retry %d (timeout)\n", attempt+1)
			continue
		}
		ackPkt, err := UnmarshalPacket(buf[:n])
		if err != nil {
			continue
		}
		if ackPkt.IsACK && ackPkt.SeqNum == s.seqNum {
			s.seqNum++
			return nil
		}
	}
	return fmt.Errorf("max retries exceeded")
}

func main() {
	fmt.Println("=== Reliable UDP (Stop-and-Wait ARQ) ===")

	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Printf("  Server error: %v\n", err)
		return
	}
	defer serverConn.Close()
	fmt.Printf("  Server on %s\n", serverConn.LocalAddr())

	var wg sync.WaitGroup

	// Receiver: accepts packets and sends ACKs
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for i := 0; i < 3; i++ {
			serverConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, remoteAddr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			pkt, err := UnmarshalPacket(buf[:n])
			if err != nil {
				continue
			}
			fmt.Printf("  Received seq=%d: %q\n", pkt.SeqNum, string(pkt.Payload))
			ack := &ReliablePacket{SeqNum: pkt.SeqNum, IsACK: true}
			serverConn.WriteToUDP(ack.Marshal(), remoteAddr)
		}
	}()

	// Sender
	clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	clientConn, _ := net.ListenUDP("udp", clientAddr)
	defer clientConn.Close()

	sender := NewReliableUDPSender(clientConn, serverConn.LocalAddr().(*net.UDPAddr))
	for _, msg := range []string{"reliable-1", "reliable-2", "reliable-3"} {
		if err := sender.Send([]byte(msg)); err != nil {
			fmt.Printf("  Send failed: %v\n", err)
		} else {
			fmt.Printf("  ACK received for: %q\n", msg)
		}
	}
	wg.Wait()
}

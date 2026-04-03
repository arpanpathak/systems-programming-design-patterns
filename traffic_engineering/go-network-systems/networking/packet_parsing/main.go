// IP and TCP header parsing from raw bytes.
package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type IPHeader struct {
	Version    uint8
	IHL        uint8
	DSCP       uint8
	ECN        uint8
	TotalLen   uint16
	ID         uint16
	Flags      uint8
	FragOffset uint16
	TTL        uint8
	Protocol   uint8
	Checksum   uint16
	SrcIP      net.IP
	DstIP      net.IP
}

func ParseIPHeader(data []byte) (*IPHeader, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("too short for IP header: %d bytes", len(data))
	}
	h := &IPHeader{
		Version:    data[0] >> 4,
		IHL:        data[0] & 0x0F,
		DSCP:       data[1] >> 2,
		ECN:        data[1] & 0x03,
		TotalLen:   binary.BigEndian.Uint16(data[2:4]),
		ID:         binary.BigEndian.Uint16(data[4:6]),
		Flags:      data[6] >> 5,
		FragOffset: binary.BigEndian.Uint16(data[6:8]) & 0x1FFF,
		TTL:        data[8],
		Protocol:   data[9],
		Checksum:   binary.BigEndian.Uint16(data[10:12]),
		SrcIP:      net.IP(data[12:16]),
		DstIP:      net.IP(data[16:20]),
	}
	return h, nil
}

func (h *IPHeader) String() string {
	proto := "unknown"
	switch h.Protocol {
	case 1:
		proto = "ICMP"
	case 6:
		proto = "TCP"
	case 17:
		proto = "UDP"
	}
	return fmt.Sprintf("IPv%d %s->%s proto=%s ttl=%d len=%d",
		h.Version, h.SrcIP, h.DstIP, proto, h.TTL, h.TotalLen)
}

type TCPHeader struct {
	SrcPort    uint16
	DstPort    uint16
	SeqNum     uint32
	AckNum     uint32
	DataOffset uint8
	Flags      TCPFlags
	Window     uint16
	Checksum   uint16
	UrgentPtr  uint16
}

type TCPFlags struct {
	FIN, SYN, RST, PSH, ACK, URG, ECE, CWR bool
}

func (f TCPFlags) String() string {
	var flags []string
	if f.SYN {
		flags = append(flags, "SYN")
	}
	if f.ACK {
		flags = append(flags, "ACK")
	}
	if f.FIN {
		flags = append(flags, "FIN")
	}
	if f.RST {
		flags = append(flags, "RST")
	}
	if f.PSH {
		flags = append(flags, "PSH")
	}
	if f.URG {
		flags = append(flags, "URG")
	}
	if len(flags) == 0 {
		return "NONE"
	}
	result := flags[0]
	for i := 1; i < len(flags); i++ {
		result += "|" + flags[i]
	}
	return result
}

func ParseTCPHeader(data []byte) (*TCPHeader, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("too short for TCP header")
	}
	h := &TCPHeader{
		SrcPort:    binary.BigEndian.Uint16(data[0:2]),
		DstPort:    binary.BigEndian.Uint16(data[2:4]),
		SeqNum:     binary.BigEndian.Uint32(data[4:8]),
		AckNum:     binary.BigEndian.Uint32(data[8:12]),
		DataOffset: data[12] >> 4,
		Flags: TCPFlags{
			FIN: data[13]&0x01 != 0, SYN: data[13]&0x02 != 0,
			RST: data[13]&0x04 != 0, PSH: data[13]&0x08 != 0,
			ACK: data[13]&0x10 != 0, URG: data[13]&0x20 != 0,
			ECE: data[13]&0x40 != 0, CWR: data[13]&0x80 != 0,
		},
		Window:    binary.BigEndian.Uint16(data[14:16]),
		Checksum:  binary.BigEndian.Uint16(data[16:18]),
		UrgentPtr: binary.BigEndian.Uint16(data[18:20]),
	}
	return h, nil
}

func (h *TCPHeader) String() string {
	return fmt.Sprintf("TCP %d->%d seq=%d ack=%d flags=[%s] win=%d",
		h.SrcPort, h.DstPort, h.SeqNum, h.AckNum, h.Flags, h.Window)
}

func main() {
	fmt.Println("=== IP/TCP Header Parsing ===")

	// Construct a fake IPv4 packet
	ipData := make([]byte, 20)
	ipData[0] = 0x45                               // v4, IHL=5
	binary.BigEndian.PutUint16(ipData[2:4], 60)    // Total length
	binary.BigEndian.PutUint16(ipData[4:6], 12345) // ID
	ipData[8] = 64                                 // TTL
	ipData[9] = 6                                  // TCP
	copy(ipData[12:16], net.ParseIP("192.168.1.100").To4())
	copy(ipData[16:20], net.ParseIP("10.0.0.1").To4())

	ip, err := ParseIPHeader(ipData)
	if err != nil {
		fmt.Printf("  IP error: %v\n", err)
		return
	}
	fmt.Printf("  %s\n", ip)

	// Construct a fake TCP header
	tcpData := make([]byte, 20)
	binary.BigEndian.PutUint16(tcpData[0:2], 54321)   // Src port
	binary.BigEndian.PutUint16(tcpData[2:4], 80)      // Dst port
	binary.BigEndian.PutUint32(tcpData[4:8], 1000)    // Seq
	tcpData[12] = 5 << 4                              // Data offset
	tcpData[13] = 0x02                                // SYN
	binary.BigEndian.PutUint16(tcpData[14:16], 65535) // Window

	tcp, err := ParseTCPHeader(tcpData)
	if err != nil {
		fmt.Printf("  TCP error: %v\n", err)
		return
	}
	fmt.Printf("  %s\n", tcp)
}

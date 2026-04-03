// Protocol concepts reference: TCP, TLS, HTTP/2, QUIC, NAT, Load Balancing.
package main

import "fmt"

func main() {
	fmt.Println("=== Protocol Concepts Reference ===")

	fmt.Println(`
  TCP 3-Way Handshake:
  ┌────────┐                      ┌────────┐
  │ Client │                      │ Server │
  └───┬────┘                      └───┬────┘
      │──── SYN (seq=x) ────────────>│
      │<─── SYN-ACK (seq=y,ack=x+1)──│
      │──── ACK (ack=y+1) ──────────>│

  TCP 4-Way Teardown:
      │──── FIN ─────────────────────>│
      │<─── ACK ──────────────────────│
      │<─── FIN ──────────────────────│
      │──── ACK ─────────────────────>│
      │   Connection Closed (TIME_WAIT)│

  TCP States: LISTEN -> SYN_RCVD -> ESTABLISHED -> FIN_WAIT_1
              -> FIN_WAIT_2 -> TIME_WAIT -> CLOSED

  TIME_WAIT: 2*MSL (~60s). Prevents old packets misinterpretation.
             SO_REUSEADDR allows binding to TIME_WAIT ports.

  TCP Congestion Control:
  ┌─────────────────────────────────────────────┐
  │ Slow Start:    cwnd doubles each RTT        │
  │ Congestion:    cwnd = ssthresh, linear grow  │
  │ Fast Retransmit: 3 duplicate ACKs           │
  │ Fast Recovery:  cwnd halved, skip slow start │
  │ Algorithms: Reno, Cubic (Linux), BBR (Google)│
  └─────────────────────────────────────────────┘

  TLS 1.3 Handshake (1-RTT):
      │──── ClientHello + KeyShare ──>│
      │<─── ServerHello + KeyShare ───│
      │<─── {EncryptedExtensions} ────│
      │<─── {Certificate} ────────────│
      │<─── {Finished} ───────────────│
      │──── {Finished} ──────────────>│

  HTTP/2 Concepts:
  ┌─────────────────────────────────────────────┐
  │ Single TCP connection, multiple streams      │
  │ Binary framing layer                         │
  │ Header compression (HPACK)                   │
  │ Server push, stream prioritization           │
  │ Flow control: per-stream and per-connection  │
  └─────────────────────────────────────────────┘

  QUIC / HTTP/3:
  ┌─────────────────────────────────────────────┐
  │ Built on UDP (avoids TCP head-of-line block) │
  │ Integrated TLS 1.3 (0-RTT resumption)       │
  │ Connection migration (via Connection IDs)    │
  │ Independent stream multiplexing              │
  └─────────────────────────────────────────────┘

  NAT:
  ┌─────────────────────────────────────────────┐
  │ SNAT: Source NAT (outbound)                  │
  │ DNAT: Destination NAT (inbound, port fwd)    │
  │ Cone NAT: Full/Restricted/Port-Restricted    │
  │ Symmetric NAT: Different mapping per dest    │
  │ Traversal: STUN, TURN, ICE                  │
  └─────────────────────────────────────────────┘

  Load Balancing:
  ┌─────────────────────────────────────────────┐
  │ L4: Round Robin, Least Connections, IP Hash  │
  │ L7: URL/Header/Cookie-based routing          │
  │ Consistent Hashing: virtual nodes, minimal   │
  │   remapping on server changes                │
  └─────────────────────────────────────────────┘
`)
}

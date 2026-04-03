// Flow tracking, traffic shaping, and network stats collection.
package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// --- Flow Tracker ---

type FlowKey struct {
	SrcIP   string
	DstIP   string
	SrcPort uint16
	DstPort uint16
	Proto   uint8
}

func (k FlowKey) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d (proto=%d)", k.SrcIP, k.SrcPort, k.DstIP, k.DstPort, k.Proto)
}

func (k FlowKey) Reverse() FlowKey {
	return FlowKey{SrcIP: k.DstIP, DstIP: k.SrcIP, SrcPort: k.DstPort, DstPort: k.SrcPort, Proto: k.Proto}
}

type FlowStats struct {
	Key       FlowKey
	Packets   int64
	Bytes     int64
	StartTime time.Time
	LastSeen  time.Time
	State     string
}

type FlowTracker struct {
	mu    sync.RWMutex
	flows map[FlowKey]*FlowStats
}

func NewFlowTracker() *FlowTracker {
	return &FlowTracker{flows: make(map[FlowKey]*FlowStats)}
}

func (ft *FlowTracker) TrackPacket(key FlowKey, size int) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	now := time.Now()

	flow, exists := ft.flows[key]
	if !exists {
		if rev, ok := ft.flows[key.Reverse()]; ok {
			rev.Packets++
			rev.Bytes += int64(size)
			rev.LastSeen = now
			if rev.State == "NEW" {
				rev.State = "ESTABLISHED"
			}
			return
		}
		flow = &FlowStats{Key: key, StartTime: now, State: "NEW"}
		ft.flows[key] = flow
	}
	flow.Packets++
	flow.Bytes += int64(size)
	flow.LastSeen = now
}

func (ft *FlowTracker) GetActiveFlows() []*FlowStats {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	var result []*FlowStats
	for _, f := range ft.flows {
		result = append(result, f)
	}
	return result
}

// --- Traffic Shaper ---

type ShapingPolicy struct {
	MaxBytesPerSecond int64
	BurstSize         int64
	tokens            int64
	lastRefill        time.Time
}

type TrafficShaper struct {
	mu       sync.Mutex
	policies map[string]*ShapingPolicy
}

func NewTrafficShaper() *TrafficShaper {
	return &TrafficShaper{policies: make(map[string]*ShapingPolicy)}
}

func (ts *TrafficShaper) AddPolicy(cidr string, maxBPS, burst int64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.policies[cidr] = &ShapingPolicy{
		MaxBytesPerSecond: maxBPS, BurstSize: burst,
		tokens: burst, lastRefill: time.Now(),
	}
}

func (ts *TrafficShaper) Allow(srcIP string, packetSize int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for cidr, policy := range ts.policies {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		ip := net.ParseIP(srcIP)
		if ip != nil && ipNet.Contains(ip) {
			now := time.Now()
			elapsed := now.Sub(policy.lastRefill).Seconds()
			policy.tokens += int64(elapsed * float64(policy.MaxBytesPerSecond))
			if policy.tokens > policy.BurstSize {
				policy.tokens = policy.BurstSize
			}
			policy.lastRefill = now
			if policy.tokens >= int64(packetSize) {
				policy.tokens -= int64(packetSize)
				return true
			}
			return false
		}
	}
	return true
}

// --- Network Stats ---

type NetworkStats struct {
	PacketsIn  int64
	BytesIn    int64
	TCPConns   int64
	UDPPackets int64
	Dropped    int64
}

func (ns *NetworkStats) RecordInbound(proto uint8, size int) {
	atomic.AddInt64(&ns.PacketsIn, 1)
	atomic.AddInt64(&ns.BytesIn, int64(size))
	switch proto {
	case 6:
		atomic.AddInt64(&ns.TCPConns, 1)
	case 17:
		atomic.AddInt64(&ns.UDPPackets, 1)
	}
}

func (ns *NetworkStats) RecordDrop() {
	atomic.AddInt64(&ns.Dropped, 1)
}

func (ns *NetworkStats) Summary() string {
	return fmt.Sprintf("Packets=%d Bytes=%d TCP=%d UDP=%d Dropped=%d",
		atomic.LoadInt64(&ns.PacketsIn), atomic.LoadInt64(&ns.BytesIn),
		atomic.LoadInt64(&ns.TCPConns), atomic.LoadInt64(&ns.UDPPackets),
		atomic.LoadInt64(&ns.Dropped))
}

func main() {
	fmt.Println("=== Flow Tracking & Traffic Shaping ===")

	// Flow tracking
	fmt.Println("\n--- Flow Tracking ---")
	tracker := NewFlowTracker()
	flows := []FlowKey{
		{SrcIP: "192.168.1.100", DstIP: "10.0.0.1", SrcPort: 54321, DstPort: 80, Proto: 6},
		{SrcIP: "192.168.1.101", DstIP: "10.0.0.2", SrcPort: 54322, DstPort: 443, Proto: 6},
		{SrcIP: "192.168.1.100", DstIP: "10.0.0.3", SrcPort: 12345, DstPort: 53, Proto: 17},
	}
	for i := 0; i < 100; i++ {
		flow := flows[rand.Intn(len(flows))]
		if rand.Float32() > 0.5 {
			flow = flow.Reverse()
		}
		tracker.TrackPacket(flow, rand.Intn(1500)+64)
	}
	for _, f := range tracker.GetActiveFlows() {
		fmt.Printf("  %s | pkts=%d bytes=%d state=%s\n", f.Key, f.Packets, f.Bytes, f.State)
	}

	// Traffic shaping
	fmt.Println("\n--- Traffic Shaping ---")
	shaper := NewTrafficShaper()
	shaper.AddPolicy("192.168.1.0/24", 10240, 5120)
	allowed, dropped := 0, 0
	for i := 0; i < 20; i++ {
		if shaper.Allow("192.168.1.100", 1024) {
			allowed++
		} else {
			dropped++
		}
	}
	fmt.Printf("  allowed=%d dropped=%d (limit: 10KB/s)\n", allowed, dropped)

	// Network stats
	fmt.Println("\n--- Network Stats ---")
	stats := &NetworkStats{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			protos := []uint8{6, 17, 1}
			stats.RecordInbound(protos[rand.Intn(len(protos))], rand.Intn(1500)+64)
			if rand.Float32() < 0.01 {
				stats.RecordDrop()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("  %s\n", stats.Summary())

	fmt.Println(`
  Traffic Concepts:
  - DPI: Deep Packet Inspection (payload beyond headers)
  - Flow-based: 5-tuple (srcIP, dstIP, srcPort, dstPort, proto)
  - QoS: DSCP marking -> EF (voice), AF (video), BE (default)
  - Queue: FIFO, WFQ, CBQ, HTB, RED
  - Apple Private Relay: 2-hop MASQUE+QUIC architecture`)
}

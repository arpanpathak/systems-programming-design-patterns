package traffic_engineering

import (
	"fmt"
	"sync"
	"time"
)

// PassiveOutlierDetection is a core Envoy traffic-shaping pattern.
// Instead of actively pinging "/health" endpoints (which adds network strain and can be faked),
// passive outlier detection monitors actual live client traffic.
// If a backend returns 5 consecutive HTTP 5xx errors, the proxy dynamically ejects it
// from the Load Balancing pool for a cooldown period.
type UpstreamNode struct {
	ID          string
	Address     string
	Active      bool
	errCount    int
	lastEjected time.Time
	mu          sync.Mutex
}

type PassiveOutlierDetector struct {
	mu           sync.RWMutex
	nodes        []*UpstreamNode
	errThreshold int           // Consecutive errors required to eject
	cooldown     time.Duration // Time required before node is allowed back in LB pool
}

func NewPassiveOutlierDetector(nodes []*UpstreamNode, threshold int, cooldown time.Duration) *PassiveOutlierDetector {
	return &PassiveOutlierDetector{
		nodes:        nodes,
		errThreshold: threshold,
		cooldown:     cooldown,
	}
}

// LogRequestResult is called inline immediately after piping a proxy response.
// If statusCode >= 500, we register a failure. If it's 200 OK, we reset the count.
func (pod *PassiveOutlierDetector) LogRequestResult(nodeID string, statusCode int) {
	pod.mu.Lock()
	defer pod.mu.Unlock()

	for _, n := range pod.nodes {
		if n.ID == nodeID {
			n.mu.Lock()
			if statusCode >= 500 {
				n.errCount++
				fmt.Printf("[Outlier Detection] Node %s returned 5xx! (Failed %d/%d)\n", n.ID, n.errCount, pod.errThreshold)

				if n.errCount >= pod.errThreshold && n.Active {
					fmt.Printf("[Outlier Detection] CRITICAL: Node %s ejected from LB Pool!\n", n.ID)
					n.Active = false
					n.lastEjected = time.Now()
				}
			} else {
				// 200 OK! Reset consecutive error count!
				n.errCount = 0
			}
			n.mu.Unlock()
			return
		}
	}
}

// GetActiveNodes returns only healthy instances for the Load Balancer to route to.
// Crucially, it automatically re-integrates ejected nodes once the cooldown expires!
func (pod *PassiveOutlierDetector) GetActiveNodes() []*UpstreamNode {
	pod.mu.RLock()
	defer pod.mu.RUnlock()

	var healthy []*UpstreamNode
	now := time.Now()

	for _, n := range pod.nodes {
		n.mu.Lock()
		if !n.Active {
			// Did it cool down?
			if now.Sub(n.lastEjected) > pod.cooldown {
				fmt.Printf("[Outlier Detection] Cooldown expired. Node %s re-integrated into LB Pool.\n", n.ID)
				n.Active = true
				n.errCount = 0 // Give it a fresh start
			}
		}

		if n.Active {
			healthy = append(healthy, n)
		}
		n.mu.Unlock()
	}

	return healthy
}

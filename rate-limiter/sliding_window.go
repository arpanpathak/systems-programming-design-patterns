package rate_limiter

import (
	"sync"
	"time"
)

// SlidingWindowLog limits traffic smoothly over a timeframe.
// Unlike traditional "Token Buckets" which permit sudden bursts, sliding windows
// offer perfectly distributed limits but consume slightly more memory tracking timestamps.
type SlidingWindowLog struct {
	mu         sync.Mutex    // Extremely critical in traffic gateways!
	timestamps []time.Time   // Slice storing exact moments requests occurred.
	windowSize time.Duration // E.g., 10 seconds.
	capacity   int           // Max requests allowed during the Window.
}

func NewSlidingWindowLog(size time.Duration, cap int) *SlidingWindowLog {
	return &SlidingWindowLog{
		timestamps: make([]time.Time, 0, cap), // Pre-allocating to avoid runtime allocations
		windowSize: size,
		capacity:   cap,
	}
}

// Allow processes an incoming API request. True = Proceed, False = Rate Limit (HTTP 429).
func (s *SlidingWindowLog) Allow() bool {
	// A high throughput ingress node will have tens of thousands of Goroutines slamming this mutex.
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	boundary := now.Add(-s.windowSize)

	// In a real optimized system, we could binary search to find the cut-off index,
	// but simple iteration from the start (oldest logs) is fast enough for localized windows.
	cutoffIdx := 0
	for i, t := range s.timestamps {
		if t.After(boundary) {
			cutoffIdx = i
			break
		}
	}

	// Fast memory shift: Slice out the ancient timestamps falling outside our window boundary!
	if cutoffIdx > 0 {
		s.timestamps = s.timestamps[cutoffIdx:]
	}

	// Make the decision
	if len(s.timestamps) < s.capacity {
		// Valid Request! We append the timestamp.
		s.timestamps = append(s.timestamps, now)
		return true
	}

	// Gateway Throttled us! Proxy must intercept and return HTTP 429 Too Many Requests.
	return false
}

package data_structures

import (
	"fmt"
	"sync"
	"time"
)

// Hierarchical Timing Wheel (Crucial Envoy/Proxy Data Structure).
// Standard Go `time.AfterFunc` allocates a timer struct and a goroutine for every single call.
// If a proxy handles 1,000,000 active connections, tracking 1 million read timeouts
// using `time.After` leads to catastrophic memory/Goroutine exhaustion.
//
// A Timing Wheel uses a fixed array (e.g., 60 slots for seconds) and a single "tick" goroutine.
// All timeouts requested for "5 seconds from now" simply append to index (CurrentTick + 5) % 60.
// This executes timeouts at scale with O(1) insertion and O(1) tick execution, avoiding OS timers entirely!
type TimingWheel struct {
	mu           sync.Mutex
	buckets      [][]func()    // Array of slices holding callbacks
	currentTick  int           // Where the clock's hand currently points
	wheelSize    int           // e.g., 60 slots
	tickDuration time.Duration // e.g., 1 Second
	stopChan     chan struct{}
}

func NewTimingWheel(wheelSize int, tickDuration time.Duration) *TimingWheel {
	return &TimingWheel{
		buckets:      make([][]func(), wheelSize),
		wheelSize:    wheelSize,
		tickDuration: tickDuration,
		stopChan:     make(chan struct{}),
	}
}

// AddTimeout places a callback in the correct future bucket linearly.
// Note: A true Hierarchical wheel handles days/hours/mins via layered wheels.
func (tw *TimingWheel) AddTimeout(delay time.Duration, callback func()) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	// Calculate how many ticks away this timeout should fire
	ticks := int(delay / tw.tickDuration)
	if ticks == 0 {
		ticks = 1 // Minimum 1 tick
	}

	targetBucket := (tw.currentTick + ticks) % tw.wheelSize
	tw.buckets[targetBucket] = append(tw.buckets[targetBucket], callback)

	fmt.Printf("[TimeWheel] Registered timeout to fire in %d ticks (Bucket %d)\n", ticks, targetBucket)
}

func (tw *TimingWheel) Start() {
	ticker := time.NewTicker(tw.tickDuration)
	go func() {
		for {
			select {
			case <-ticker.C:
				tw.advanceClock()
			case <-tw.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (tw *TimingWheel) Stop() {
	close(tw.stopChan)
}

func (tw *TimingWheel) advanceClock() {
	tw.mu.Lock()
	// Move the clock hand forward
	tw.currentTick = (tw.currentTick + 1) % tw.wheelSize

	// Grab all callbacks waiting in this specific second
	callbacksToFire := tw.buckets[tw.currentTick]
	// Empty out the bucket for the next revolution
	tw.buckets[tw.currentTick] = nil
	tw.mu.Unlock()

	// Execute them concurrently or sequentially based on proxy rules
	if len(callbacksToFire) > 0 {
		fmt.Printf("[TimeWheel] Tick! Firing %d scheduled timeouts from Bucket %d...\n", len(callbacksToFire), tw.currentTick)
		for _, cb := range callbacksToFire {
			go cb() // Execute outside the lock
		}
	}
}

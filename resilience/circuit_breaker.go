package resilience

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker acts as a state machine that guards downstream microservices.
// This prevents catastrophic cascading failures in a Mesh infrastructure.
// Apple asks this because Envoy proxies use Circuit Breakers heavily.
type State int

const (
	StateClosed   State = iota // Healthy! Traffic flows normally.
	StateOpen                  // Total failure! Drop all traffic to allow the upstream to recover.
	StateHalfOpen              // Testing! Allow 1 request through to see if the upstream recovered.
)

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	threshold    int           // Allowable failure count before Tripping Circuit OPEN
	timeout      time.Duration // Time to wait in OPEN state before trying HALF-OPEN
	lastFailTime time.Time     // Critical for tracking timeouts
}

var ErrCircuitOpen = errors.New("circuit breaker is OPEN. traffic dropped fast-fail")

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

// Allow is called by the routing engine BEFORE forwarding the proxy request.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil // Go ahead!
	case StateOpen:
		// Has enough time passed since the proxy tripped?
		if time.Since(cb.lastFailTime) > cb.timeout {
			fmt.Println("[CB] State transitioned from OPEN -> HALF-OPEN. Testing recovery...")
			cb.state = StateHalfOpen
			return nil // Let exactly ONE request slip through!
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		// If another concurrent request tries while we are testing Half-Open, drop it. We only need 1 sample.
		return ErrCircuitOpen
	}
	return nil
}

// RecordResult must be called after the network request is initiated downstream.
func (cb *CircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		fmt.Printf("[CB] Registered Failure! Error: %v\n", err)
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == StateHalfOpen || (cb.state == StateClosed && cb.failures >= cb.threshold) {
			fmt.Println("\n[CB] *** TRIPPED! Circuit is now OPEN. Fast-failing future traffic! ***")
			cb.state = StateOpen
		}
	} else {
		// Recovery path
		cb.failures = 0
		if cb.state == StateHalfOpen {
			fmt.Println("\n[CB] Request Succeeded! State transitioned HALF-OPEN -> CLOSED.")
			cb.state = StateClosed
		}
	}
}

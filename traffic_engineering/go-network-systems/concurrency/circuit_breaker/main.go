// Circuit Breaker: protect against cascading failures.
// States: CLOSED (normal) -> OPEN (failing) -> HALF-OPEN (probing).
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	}
	return "UNKNOWN"
}

type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailure      time.Time
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is OPEN")
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
		return err
	}

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	} else {
		cb.failureCount = 0
	}
	return nil
}

func main() {
	fmt.Println("=== Circuit Breaker ===")

	cb := NewCircuitBreaker(3, 2, 200*time.Millisecond)

	callCount := 0
	unreliableService := func() error {
		callCount++
		if callCount <= 4 {
			return errors.New("service unavailable")
		}
		return nil
	}

	for i := 0; i < 10; i++ {
		err := cb.Execute(unreliableService)
		cb.mu.Lock()
		state := cb.state
		cb.mu.Unlock()
		if err != nil {
			fmt.Printf("  Call %d: ERROR(%v) State=%s\n", i, err, state)
		} else {
			fmt.Printf("  Call %d: SUCCESS State=%s\n", i, state)
		}
		time.Sleep(80 * time.Millisecond)
	}
}

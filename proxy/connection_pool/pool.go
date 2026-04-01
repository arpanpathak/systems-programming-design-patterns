package connection_pool

import (
	"errors"
	"fmt"
	"sync"
)

// A Connection represents an expensive TCP handshake to a backend service.
// (In reality, this would wrap net.Conn)
type Connection struct {
	ID int
	// ... metadata, socket FDs
}

// ConnectionPool manages reusing alive connections instead of dialing new ones
// for every proxy request. If all connections are currently in use, it forces
// new proxy threads to BLOCK AND WAIT for an existing connection to be returned!
type ConnectionPool struct {
	mu          sync.Mutex // Guards the internal state
	cond        *sync.Cond // Signals waiting Goroutines when a connection becomes free!
	queue       []*Connection
	activeCount int // How many connections are currently out being used
	maxCapacity int // Hard limit of allowable open sockets
}

func NewConnectionPool(capacity int) *ConnectionPool {
	cp := &ConnectionPool{
		maxCapacity: capacity,
	}
	// sync.Cond is tied to our primary Mutex
	cp.cond = sync.NewCond(&cp.mu)
	return cp
}

// Acquire requests a connection from the pool.
// If the pool is maxed out, it BLOCKS using sync.Cond.Wait()—which puts the goroutine
// to sleep (releasing CPU) until another goroutine calls Release().
func (cp *ConnectionPool) Acquire() (*Connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// While loop is required for sync.Cond to prevent "Spurious Wakeups"
	for len(cp.queue) == 0 && cp.activeCount >= cp.maxCapacity {
		fmt.Println("Pool exhausted! Putting Goroutine to sleep until `Wait` is notified...")
		// Wait atomically unlocks `cp.mu`, sleeps, and re-locks `cp.mu` when woken!
		cp.cond.Wait()
	}

	// Case 1: We have idle connections in our queue slice. Re-use them!
	if len(cp.queue) > 0 {
		conn := cp.queue[0]
		cp.queue = cp.queue[1:] // Pop front
		cp.activeCount++
		fmt.Printf("Re-using Idle Connection [%d]\n", conn.ID)
		return conn, nil
	}

	// Case 2: Pool isn't full, but we have no idle connections. Provision a brand new one!
	if cp.activeCount < cp.maxCapacity {
		cp.activeCount++
		conn := &Connection{ID: cp.activeCount}
		fmt.Printf("Dialing Brand New Upstream Connection [%d]\n", conn.ID)
		return conn, nil
	}

	return nil, errors.New("unreachable pool state")
}

// Release returns the socket to the idle queue for another request to use.
func (cp *ConnectionPool) Release(conn *Connection) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.queue = append(cp.queue, conn)
	cp.activeCount--

	fmt.Printf("Returned Connection [%d] back to the Pool queue!\n", conn.ID)
	// WAKE UP exactly ONE waiting Goroutine stuck in `Acquire()`
	// (Envoy uses this exact Cond logic for connection pacing).
	cp.cond.Signal()
}

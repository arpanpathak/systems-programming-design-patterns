// Reactor Pattern: single-threaded event dispatch loop.
package main

import (
	"fmt"
	"sync"
	"time"
)

type EventType string

const (
	EventRead  EventType = "READ"
	EventWrite EventType = "WRITE"
	EventError EventType = "ERROR"
)

type Event struct {
	Type    EventType
	Payload interface{}
}

type EventHandler func(Event)

type Reactor struct {
	handlers map[EventType][]EventHandler
	events   chan Event
	quit     chan struct{}
	mu       sync.RWMutex
}

func NewReactor(bufSize int) *Reactor {
	return &Reactor{
		handlers: make(map[EventType][]EventHandler),
		events:   make(chan Event, bufSize),
		quit:     make(chan struct{}),
	}
}

func (r *Reactor) Register(eventType EventType, handler EventHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[eventType] = append(r.handlers[eventType], handler)
}

func (r *Reactor) Emit(event Event) {
	select {
	case r.events <- event:
	case <-r.quit:
	}
}

func (r *Reactor) Run() {
	for {
		select {
		case <-r.quit:
			return
		case event := <-r.events:
			r.mu.RLock()
			handlers := r.handlers[event.Type]
			r.mu.RUnlock()
			for _, h := range handlers {
				h(event)
			}
		}
	}
}

func (r *Reactor) Stop() { close(r.quit) }

func main() {
	fmt.Println("=== Reactor Pattern ===")

	reactor := NewReactor(100)

	reactor.Register(EventRead, func(e Event) {
		fmt.Printf("  READ handler: %v\n", e.Payload)
	})
	reactor.Register(EventWrite, func(e Event) {
		fmt.Printf("  WRITE handler: %v\n", e.Payload)
	})
	reactor.Register(EventError, func(e Event) {
		fmt.Printf("  ERROR handler: %v\n", e.Payload)
	})

	go reactor.Run()

	reactor.Emit(Event{Type: EventRead, Payload: "data chunk 1"})
	reactor.Emit(Event{Type: EventWrite, Payload: "flush buffer"})
	reactor.Emit(Event{Type: EventRead, Payload: "data chunk 2"})
	reactor.Emit(Event{Type: EventError, Payload: "connection reset"})

	time.Sleep(100 * time.Millisecond)
	reactor.Stop()
}

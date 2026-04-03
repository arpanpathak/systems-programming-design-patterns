// Graceful Shutdown + ErrGroup pattern.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// Graceful Shutdown
// =============================================================================

type Server struct {
	quit    chan struct{}
	done    chan struct{}
	handler func(req int) string
}

func NewServer(handler func(int) string) *Server {
	return &Server{
		quit:    make(chan struct{}),
		done:    make(chan struct{}),
		handler: handler,
	}
}

func (s *Server) Start() {
	go func() {
		defer close(s.done)
		reqID := 0
		for {
			select {
			case <-s.quit:
				fmt.Println("  Server: draining remaining requests...")
				time.Sleep(50 * time.Millisecond)
				fmt.Println("  Server: shutdown complete")
				return
			default:
				reqID++
				if reqID > 5 {
					time.Sleep(20 * time.Millisecond)
					continue
				}
				result := s.handler(reqID)
				fmt.Printf("  Server handled request %d: %s\n", reqID, result)
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.quit)
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// =============================================================================
// ErrGroup: run goroutines, cancel all on first error.
// =============================================================================

type ErrGroup struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewErrGroup(ctx context.Context) *ErrGroup {
	ctx, cancel := context.WithCancel(ctx)
	return &ErrGroup{ctx: ctx, cancel: cancel}
}

func (g *ErrGroup) Go(fn func(ctx context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *ErrGroup) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

func main() {
	fmt.Println("=== Graceful Shutdown ===")
	srv := NewServer(func(id int) string {
		return fmt.Sprintf("response-%d", id)
	})
	srv.Start()
	time.Sleep(150 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("  Shutdown error:", err)
	} else {
		fmt.Println("  Shutdown successful")
	}

	fmt.Println("\n=== ErrGroup Pattern ===")
	eg := NewErrGroup(context.Background())

	eg.Go(func(ctx context.Context) error {
		fmt.Println("  Task 1: starting")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  Task 1: done")
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		fmt.Println("  Task 2: starting")
		time.Sleep(30 * time.Millisecond)
		return errors.New("task 2 failed")
	})
	eg.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			fmt.Println("  Task 3: cancelled due to peer failure")
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			fmt.Println("  Task 3: done")
			return nil
		}
	})

	if err := eg.Wait(); err != nil {
		fmt.Println("  ErrGroup failed:", err)
	}
}

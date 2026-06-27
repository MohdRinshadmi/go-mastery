// Day 13 walkthrough — context, mutex, atomic. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Mutex-guarded counter — safe under -race.
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// doWork respects cancellation via ctx.Done().
func doWork(ctx context.Context, d time.Duration) (string, error) {
	select {
	case <-time.After(d):
		return "completed", nil
	case <-ctx.Done():
		return "", ctx.Err() // DeadlineExceeded or Canceled
	}
}

func main() {
	fmt.Println("== Mutex counter (1000 goroutines) ==")
	var wg sync.WaitGroup
	c := &Counter{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println("  final (want 1000):", c.Value())

	fmt.Println("== atomic counter (lock-free) ==")
	var a atomic.Int64
	var wg2 sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() { defer wg2.Done(); a.Add(1) }()
	}
	wg2.Wait()
	fmt.Println("  final (want 1000):", a.Load())

	fmt.Println("== context timeout ==")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	// work takes 200ms but ctx times out at 50ms -> cancelled
	if _, err := doWork(ctx, 200*time.Millisecond); err != nil {
		fmt.Println("  slow work cancelled:", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()
	if res, err := doWork(ctx2, 10*time.Millisecond); err == nil {
		fmt.Println("  fast work:", res)
	}
}

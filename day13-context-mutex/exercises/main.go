// Day 13 — YOUR exercises. Run with the race detector: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =====================================================================
// EXERCISE 1 — fix the race
// SafeCounter.Inc currently has NO synchronization. Running the program
// under -race will report a data race. Add a sync.Mutex (or atomic) so
// `go run -race main.go` is CLEAN and the final value is exactly 1000.
// =====================================================================

type SafeCounter struct {
	// TODO: add a mutex
	n int
}

func (c *SafeCounter) Inc() {
	// TODO: synchronize this
	c.n++
}
func (c *SafeCounter) Value() int {
	// TODO: synchronize this
	return c.n
}

// =====================================================================
// EXERCISE 2 — context timeout
// Implement fetchWithTimeout: simulate work that takes `work` duration,
// but return ctx.Err() if the context is cancelled/expires first.
// Use select with time.After(work) and ctx.Done().
// =====================================================================

func fetchWithTimeout(ctx context.Context, work time.Duration) (string, error) {
	// TODO
	return "", nil
}

// =====================================================================
// CHALLENGE — concurrency-safe cache (map + RWMutex)
// Implement Cache with Get(key) (string, bool) and Set(key, value).
// Must pass `go test -race` if you wrote tests; at minimum, hammering it
// from many goroutines under -race must be clean.
// =====================================================================

type Cache struct {
	// TODO: sync.RWMutex + map[string]string
}

func NewCache() *Cache {
	// TODO
	return &Cache{}
}
func (c *Cache) Get(key string) (string, bool) {
	// TODO (RLock)
	return "", false
}
func (c *Cache) Set(key, value string) {
	// TODO (Lock)
}

func main() {
	fmt.Println("== Exercise 1: SafeCounter under -race ==")
	var wg sync.WaitGroup
	c := &SafeCounter{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println("  final (want 1000):", c.Value())

	fmt.Println("== Exercise 2: fetchWithTimeout ==")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := fetchWithTimeout(ctx, 100*time.Millisecond)
	fmt.Println("  err (want deadline exceeded):", err)

	fmt.Println("== Challenge: concurrent Cache ==")
	// TODO: spin up goroutines doing Set/Get concurrently, run under -race
}

// Day 13 — reference solutions. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SafeCounter struct {
	mu sync.Mutex
	n  int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}
func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func fetchWithTimeout(ctx context.Context, work time.Duration) (string, error) {
	select {
	case <-time.After(work):
		return "data", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type Cache struct {
	mu sync.RWMutex
	m  map[string]string
}

func NewCache() *Cache { return &Cache{m: make(map[string]string)} }

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
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

	fmt.Println("== Challenge: concurrent Cache under -race ==")
	cache := NewCache()
	var wg2 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg2.Add(2)
		key := fmt.Sprintf("k%d", i%10)
		go func() { defer wg2.Done(); cache.Set(key, "v") }()
		go func() { defer wg2.Done(); cache.Get(key) }()
	}
	wg2.Wait()
	v, ok := cache.Get("k1")
	fmt.Printf("  Get(k1) = %q, %v\n", v, ok)
}

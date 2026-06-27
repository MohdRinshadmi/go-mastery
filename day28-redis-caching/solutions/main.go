// Day 28 walkthrough — cache-aside + singleflight stampede protection.
// Run: go run .
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// ---- slow "database" with a call counter -------------------------------
type DB struct{ calls atomic.Int64 }

func (d *DB) Get(_ context.Context, key string) (string, error) {
	d.calls.Add(1)
	time.Sleep(50 * time.Millisecond) // simulate a slow query
	return "value-of-" + key, nil
}

// ---- cache-aside with TTL + singleflight -------------------------------
type entry struct {
	val string
	exp time.Time
}
type Cache struct {
	db    *DB
	ttl   time.Duration
	mu    sync.Mutex
	store map[string]entry
	group singleflight.Group // collapses concurrent misses into 1 call
	Hits  atomic.Int64
	Miss  atomic.Int64
}

func NewCache(db *DB, ttl time.Duration) *Cache {
	return &Cache{db: db, ttl: ttl, store: map[string]entry{}}
}

func (c *Cache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.store[key]
	if !ok || time.Now().After(e.exp) {
		return "", false
	}
	return e.val, true
}
func (c *Cache) set(key, val string) {
	c.mu.Lock()
	c.store[key] = entry{val: val, exp: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	if v, ok := c.get(key); ok {
		c.Hits.Add(1)
		return v, nil
	}
	c.Miss.Add(1)
	// singleflight: only ONE goroutine runs the loader per key; the rest
	// wait for and share its result -> stampede protection.
	v, err, _ := c.group.Do(key, func() (any, error) {
		val, err := c.db.Get(ctx, key)
		if err != nil {
			return "", err
		}
		c.set(key, val)
		return val, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func main() {
	db := &DB{}
	cache := NewCache(db, time.Minute)
	ctx := context.Background()

	fmt.Println("== 50 concurrent gets for the SAME cold key ==")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = cache.Get(ctx, "p1") }()
	}
	wg.Wait()
	fmt.Printf("  DB calls=%d (singleflight collapsed 50 misses into 1)\n", db.calls.Load())

	fmt.Println("== repeated gets hit the cache ==")
	for i := 0; i < 5; i++ {
		_, _ = cache.Get(ctx, "p1")
	}
	fmt.Printf("  DB calls=%d hits=%d misses=%d\n", db.calls.Load(), cache.Hits.Load(), cache.Miss.Load())
}

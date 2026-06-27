// Day 28 debugging — cache stampede (thundering herd) on a hot key.
//
// We simulate Redis with an in-memory map and Postgres with a slow function.
// The cache-aside Get is correct for ONE caller: check cache, on miss load from
// DB, populate, return. But under concurrency it has no stampede protection:
// when a hot key is cold (first request, or just after expiry), N goroutines all
// miss at once, all see "not cached", and all hammer the slow DB in parallel.
//
// In production that's the moment your database falls over: one popular product
// expires and 10,000 requests hit Postgres simultaneously.
//
// Stdlib only. Run with the race detector:  go run -race .
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type cache struct {
	mu sync.Mutex
	m  map[string]string
}

func newCache() *cache { return &cache{m: map[string]string{}} }

func (c *cache) get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[k]
	return v, ok
}

func (c *cache) set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
}

// db is the slow source of truth. dbCalls counts how often we actually hit it.
type db struct {
	dbCalls atomic.Int64
}

func (d *db) load(key string) string {
	d.dbCalls.Add(1)
	time.Sleep(50 * time.Millisecond) // expensive query
	return "product:" + key
}

type store struct {
	c  *cache
	db *db
}

// Get is cache-aside with NO stampede protection.
func (s *store) Get(key string) string {
	if v, ok := s.c.get(key); ok {
		return v // hit
	}
	// BUG: every concurrent miss reaches here and independently loads the DB.
	v := s.db.load(key)
	s.c.set(key, v)
	return v
}

func main() {
	s := &store{c: newCache(), db: &db{}}

	const concurrent = 50
	var wg sync.WaitGroup
	wg.Add(concurrent)
	// 50 requests for the SAME cold hot-key, all at once.
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			_ = s.Get("hot-key")
		}()
	}
	wg.Wait()

	calls := s.db.dbCalls.Load()
	fmt.Printf("DB calls for one hot key under %d concurrent requests: %d\n", concurrent, calls)
	if calls > 1 {
		fmt.Printf("BUG: cache stampede — expected 1 DB call, got %d\n", calls)
	} else {
		fmt.Println("OK: collapsed to a single DB call")
	}
}

// Day 28 debugging — FIXED: single-flight collapses concurrent misses.
//
// Same cache-aside, but we add an in-process single-flight: the FIRST goroutine
// to miss a key starts the DB load; every other goroutine that misses the SAME
// key while that load is in flight WAITS for its result instead of starting its
// own DB call. N concurrent misses for one key become exactly ONE DB call.
//
// In production you'd use golang.org/x/sync/singleflight (and a Redis-level lock
// for the cross-process case). This hand-rolled version keeps it stdlib-only and
// shows exactly what singleflight does.
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

type db struct {
	dbCalls atomic.Int64
}

func (d *db) load(key string) string {
	d.dbCalls.Add(1)
	time.Sleep(50 * time.Millisecond)
	return "product:" + key
}

// call tracks one in-flight load that other callers can wait on.
type call struct {
	wg  sync.WaitGroup
	val string
}

type store struct {
	c  *cache
	db *db

	mu      sync.Mutex
	inFlight map[string]*call // key -> in-progress load
}

func newStore() *store {
	return &store{c: newCache(), db: &db{}, inFlight: map[string]*call{}}
}

// Get is cache-aside WITH single-flight stampede protection.
func (s *store) Get(key string) string {
	if v, ok := s.c.get(key); ok {
		return v // hit
	}

	s.mu.Lock()
	if cl, ok := s.inFlight[key]; ok {
		// Someone is already loading this key — wait for their result.
		s.mu.Unlock()
		cl.wg.Wait()
		return cl.val
	}
	// We're the first miss: register an in-flight call and do the load.
	cl := &call{}
	cl.wg.Add(1)
	s.inFlight[key] = cl
	s.mu.Unlock()

	cl.val = s.db.load(key)
	s.c.set(key, cl.val)

	s.mu.Lock()
	delete(s.inFlight, key)
	s.mu.Unlock()
	cl.wg.Done() // release the waiters

	return cl.val
}

func main() {
	s := newStore()

	const concurrent = 50
	var wg sync.WaitGroup
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			_ = s.Get("hot-key")
		}()
	}
	wg.Wait()

	calls := s.db.dbCalls.Load()
	fmt.Printf("DB calls for one hot key under %d concurrent requests: %d\n", concurrent, calls)
	if calls == 1 {
		fmt.Println("OK: single-flight collapsed 50 misses into 1 DB call")
	} else {
		fmt.Printf("unexpected: got %d DB calls\n", calls)
	}
}

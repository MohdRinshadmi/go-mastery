// Day 23 debugging — FIXED.
//
// Guard the counter map with a mutex so concurrent increments are safe and
// exact. (The real prometheus client does this for you with atomics; here we
// do it by hand to show what "thread-safe metric" actually means.)
//
// Bonus fix for the lesson's #1 metrics sin — CARDINALITY: we normalize the
// label to a bounded ROUTE TEMPLATE before recording, so /orders/123 and
// /orders/456 both record as /orders/{id} instead of creating a new time
// series per ID.
//
// Run with: go run -race .
package main

import (
	"fmt"
	"strings"
	"sync"
)

type counterVec struct {
	mu     sync.Mutex
	values map[string]int64
}

func newCounterVec() *counterVec {
	return &counterVec{values: make(map[string]int64)}
}

// Inc increments the counter for a label value, safely.
func (c *counterVec) Inc(label string) {
	c.mu.Lock()
	c.values[label]++
	c.mu.Unlock()
}

func (c *counterVec) Get(label string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.values[label]
}

// templatePath collapses high-cardinality paths to a bounded template so a
// numeric ID segment doesn't spawn a new time series per request.
func templatePath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if s != "" && strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
			segs[i] = "{id}"
		}
	}
	return strings.Join(segs, "/")
}

func main() {
	const goroutines = 50
	const perGoroutine = 1000

	requests := newCounterVec()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			// Each goroutine hits a DIFFERENT numeric order id; templating
			// keeps them all under one bounded series.
			path := templatePath(fmt.Sprintf("/orders/%d", g))
			for i := 0; i < perGoroutine; i++ {
				requests.Inc(path)
			}
		}(g)
	}
	wg.Wait()

	want := int64(goroutines * perGoroutine)
	got := requests.Get("/orders/{id}")
	fmt.Printf("http_requests_total{path=%q} = %d (want %d)\n", "/orders/{id}", got, want)
	if got == want {
		fmt.Println("=> exact count, one bounded series: race-free and cardinality-safe")
	}
}

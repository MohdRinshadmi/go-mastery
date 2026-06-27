// Day 23 debugging — a metrics counter with a DATA RACE.
//
// We simulate a Prometheus-style CounterVec with a plain map (STDLIB ONLY,
// no client_golang). A middleware increments http_requests_total{path} from
// every request handler. Under concurrent load the counter is read and
// written from many goroutines with no synchronization — a data race that
// (a) trips `go run -race .` and (b) silently undercounts requests, so your
// dashboards lie about traffic.
//
// Run with: go run -race .
package main

import (
	"fmt"
	"sync"
)

// counterVec mimics prometheus.CounterVec: a counter per label value.
type counterVec struct {
	values map[string]int64
}

func newCounterVec() *counterVec {
	return &counterVec{values: make(map[string]int64)}
}

// Inc increments the counter for the given label value.
//
// BUG: map read + write with zero synchronization. Concurrent calls race on
// both the map header (can panic: "concurrent map writes") and the int64
// (lost updates → undercount).
func (c *counterVec) Inc(label string) {
	c.values[label]++
}

func (c *counterVec) Get(label string) int64 {
	return c.values[label]
}

func main() {
	const goroutines = 50
	const perGoroutine = 1000
	const path = "/orders"

	requests := newCounterVec()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				requests.Inc(path) // concurrent unsynchronized writes
			}
		}()
	}
	wg.Wait()

	want := int64(goroutines * perGoroutine)
	got := requests.Get(path)
	fmt.Printf("http_requests_total{path=%q} = %d (want %d)\n", path, got, want)
	if got != want {
		fmt.Println("=> metric UNDERCOUNTED due to the data race; dashboards are wrong")
	}
}

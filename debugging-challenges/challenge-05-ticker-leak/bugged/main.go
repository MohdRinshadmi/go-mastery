package main

import (
	"fmt"
	"runtime"
	"time"
)

func doRefresh() int { return 1 }

// startPoller launches a background ticker goroutine for a request.
// BUG: the ticker is never stopped, and the goroutine has no exit path, so
// `for range ticker.C` blocks forever. One leaked goroutine + ticker per call.
func startPoller() {
	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		for range ticker.C {
			_ = doRefresh()
		}
	}()
	// no Stop(), no done channel — nothing tears this down
}

// handleRequest simulates one short-lived request that uses a poller.
func handleRequest() {
	startPoller()
	time.Sleep(5 * time.Millisecond) // "do the request work"
	// request returns... but the poller goroutine and ticker live on forever
}

func main() {
	runtime.GC()
	before := runtime.NumGoroutine()
	fmt.Printf("goroutines before: %d\n", before)

	for i := 0; i < 100; i++ {
		handleRequest()
	}

	// give any goroutines that COULD exit a chance to, then measure.
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()

	fmt.Printf("goroutines after:  %d\n", after)
	fmt.Printf("leaked:            %d (want ~0)\n", after-before)
}

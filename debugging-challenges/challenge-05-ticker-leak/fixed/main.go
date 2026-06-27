package main

import (
	"fmt"
	"runtime"
	"time"
)

func doRefresh() int { return 1 }

// startPoller returns a stop func. The goroutine exits on `done`, and the
// ticker is stopped via defer. Net leak per call: zero.
func startPoller() (stop func()) {
	ticker := time.NewTicker(10 * time.Millisecond)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop() // FIX: release the runtime timer
		for {
			select {
			case <-ticker.C:
				_ = doRefresh()
			case <-done: // FIX: explicit exit so the goroutine returns
				return
			}
		}
	}()

	return func() { close(done) }
}

func handleRequest() {
	stop := startPoller()
	defer stop() // guaranteed teardown on every path

	time.Sleep(5 * time.Millisecond)
}

func main() {
	runtime.GC()
	before := runtime.NumGoroutine()
	fmt.Printf("goroutines before: %d\n", before)

	for i := 0; i < 100; i++ {
		handleRequest()
	}

	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()

	fmt.Printf("goroutines after:  %d\n", after)
	fmt.Printf("leaked:            %d (want ~0)\n", after-before)
}

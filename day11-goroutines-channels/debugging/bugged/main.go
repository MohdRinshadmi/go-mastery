// Day 11 debugging — BUGGED.
//
// Scenario: search several "replicas" for an answer and return the FIRST one.
// We launch one goroutine per replica, each sends its result on a shared
// UNBUFFERED channel. We read exactly one value (the winner) and return.
//
// Bug: the losing goroutines are still blocked on `out <- ...` forever, because
// nobody ever receives their value (we only read one) and the channel is
// unbuffered. Each call to firstReplica leaks N-1 goroutines.
//
// We prove the leak with runtime.NumGoroutine(): it climbs every call and
// never comes back down.
package main

import (
	"fmt"
	"runtime"
	"time"
)

// firstReplica asks `replicas` workers for an answer and returns the first.
func firstReplica(replicas int) int {
	out := make(chan int) // UNBUFFERED: a send blocks until someone receives

	for r := 0; r < replicas; r++ {
		go func(id int) {
			// Pretend each replica does some work, then reports.
			time.Sleep(time.Duration(id) * time.Millisecond)
			out <- id // BUG: only the winner is ever received; the rest block here forever
		}(r)
	}

	return <-out // take the first, ignore (and strand) the rest
}

func main() {
	const replicas = 5
	const rounds = 20

	for i := 0; i < rounds; i++ {
		_ = firstReplica(replicas)
	}

	// Give the runtime a moment so stragglers are parked, then count.
	time.Sleep(100 * time.Millisecond)

	leaked := runtime.NumGoroutine() - 1 // minus main
	fmt.Printf("after %d rounds of %d replicas: %d goroutines still alive\n", rounds, replicas, leaked)

	// Expected if correct: ~0. We leak (replicas-1) per round.
	if leaked > replicas {
		fmt.Printf("GOROUTINE LEAK detected: %d stranded goroutines blocked on `out <- id`\n", leaked)
	}
}

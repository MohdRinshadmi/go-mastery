// Day 11 debugging — FIXED.
//
// Same "first replica wins" scenario, but no goroutine leaks.
//
// Fix: give the channel enough buffer that EVERY sender can complete its send
// even though we only receive one value. With `make(chan int, replicas)` each
// of the N goroutines lands its value in the buffer and returns — the losers
// don't block, so they don't leak.
//
// (Other valid fixes: a `done` channel + select on send, or context cancellation
// — see Day 13/15. Buffering is the simplest correct fix for a fixed, known
// number of one-shot senders.)
package main

import (
	"fmt"
	"runtime"
	"time"
)

func firstReplica(replicas int) int {
	out := make(chan int, replicas) // buffered: every sender can complete

	for r := 0; r < replicas; r++ {
		go func(id int) {
			time.Sleep(time.Duration(id) * time.Millisecond)
			out <- id // never blocks — there is always a free buffer slot
		}(r)
	}

	return <-out // take the first; the others drop their value into the buffer and exit
}

func main() {
	const replicas = 5
	const rounds = 20

	for i := 0; i < rounds; i++ {
		_ = firstReplica(replicas)
	}

	time.Sleep(100 * time.Millisecond)

	leaked := runtime.NumGoroutine() - 1 // minus main
	fmt.Printf("after %d rounds of %d replicas: %d goroutines still alive\n", rounds, replicas, leaked)

	if leaked <= 1 {
		fmt.Println("no leak: every replica goroutine completed its send and exited")
	}
}

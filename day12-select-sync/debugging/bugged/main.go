// Day 12 debugging — BUGGED.
//
// Scenario: run N tasks concurrently, wait for all of them with a WaitGroup,
// then report the total. (We repeat the experiment for a few rounds so the
// timing-dependent bug shows up reliably on a single run.)
//
// Bug: `wg.Add(1)` is called INSIDE each goroutine instead of before `go`.
// There is a race between Add and Wait: the loop can launch all goroutines and
// reach wg.Wait() before any of them has run its wg.Add(1). At that instant the
// counter is 0, so Wait() returns immediately — we "finish" while tasks are
// still running, and the total is short.
//
// The symptom is the *opposite* of a deadlock: Wait returns too early. We make
// it visible by comparing `completed` against the expected total. The program
// always exits promptly (no hang). Both `go vet` and `go run -race .` flag the
// misuse — `go vet`: "WaitGroup.Add called from inside new goroutine".
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func runRound(tasks int) int64 {
	var wg sync.WaitGroup
	var completed atomic.Int64

	for i := 0; i < tasks; i++ {
		go func(id int) {
			wg.Add(1) // BUG: must be called BEFORE `go`, not inside the goroutine
			defer wg.Done()
			completed.Add(1)
		}(i)
	}

	wg.Wait() // may return while completed < tasks because Add hasn't run yet
	return completed.Load()
}

func main() {
	const tasks = 50
	const rounds = 8

	worst := int64(tasks)
	early := 0
	for r := 0; r < rounds; r++ {
		got := runRound(tasks)
		if got < worst {
			worst = got
		}
		if got != int64(tasks) {
			early++
		}
	}

	fmt.Printf("ran %d rounds of %d tasks; worst round completed only %d/%d before Wait returned\n",
		rounds, tasks, worst, tasks)
	if early > 0 {
		fmt.Printf("BUG: WaitGroup returned EARLY in %d/%d rounds (Add ran after Wait)\n", early, rounds)
	}
}

// Day 12 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// Mentor review: I'll look for correct WaitGroup usage (Add before go, defer Done),
// proper timeout patterns, and no goroutine leaks.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import (
	"fmt"
	"sync"
	"time"
)

// =====================================================================
// EXERCISE 1 (beginner) — Buffered collect
//
// Implement collectResults(n int) []string that:
//   - Launches n goroutines, each computing fmt.Sprintf("result-%d", id)
//     after sleeping id*5 milliseconds.
//   - Uses a BUFFERED channel of capacity n to collect all results
//     without any goroutine ever blocking.
//   - Returns the collected slice (order doesn't matter).
//
// Constraint: do NOT use sync.WaitGroup. Use only the channel.
// =====================================================================

func collectResults(n int) []string {
	// TODO: implement
	return nil
}

// =====================================================================
// EXERCISE 2 (beginner) — First wins with select + timeout
//
// Implement firstOf(sources []<-chan string, timeout time.Duration) (string, bool)
// that returns the FIRST value received from ANY of the sources channels,
// and true. If nothing arrives within timeout, return "", false.
//
// HINT: you cannot have a variable number of cases in a select statement
// directly. You'll need to merge the sources into one channel first.
// Implement mergeStrings(sources []<-chan string) <-chan string as a helper
// that fans all sources into one channel.
// =====================================================================

func mergeStrings(sources []<-chan string) <-chan string {
	// TODO: implement. Each source needs its own goroutine reading from it.
	// The merged channel should be closed when ALL sources are done.
	// HINT: use a WaitGroup.
	return nil
}

func firstOf(sources []<-chan string, timeout time.Duration) (string, bool) {
	// TODO: use mergeStrings + select + time.After
	return "", false
}

// =====================================================================
// EXERCISE 3 (beginner) — Parallel map with WaitGroup
//
// Implement parallelMap(inputs []int, fn func(int) int) []int that:
//   - Applies fn to each element concurrently (one goroutine per element).
//   - Preserves ORDER: result[i] == fn(inputs[i]).
//   - Uses sync.WaitGroup (NOT channels) for synchronization.
//   - Returns the transformed slice.
//
// This is safe because each goroutine writes to a different index.
// =====================================================================

func parallelMap(inputs []int, fn func(int) int) []int {
	// TODO: implement. Pre-allocate results slice, then launch goroutines.
	return nil
}

// =====================================================================
// CHALLENGE (intermediate) — Rate-limited worker with done signal
//
// Implement rateWorker(jobs <-chan int, rate int, done <-chan struct{}) <-chan int
// that:
//   - Reads from jobs channel.
//   - Processes AT MOST `rate` jobs per second (use a ticker).
//   - Stops cleanly when done is closed.
//   - Closes its output channel when it stops.
//   - For each job processed, sends job*2 on the output channel.
//
// In main: create a jobs channel with 10 jobs, launch rateWorker with
// rate=5 (5/sec), let it run for 300ms, then close done and drain output.
// Print how many jobs were processed. (Expect around 1-2 at 5/sec in 300ms)
//
// HINT: time.NewTicker(time.Second / time.Duration(rate)) gives you a tick
// every 1/rate seconds. Each tick = permission to process one job.
// Use select with: jobs case, ticker.C case (acquire permission), done case.
// =====================================================================

func rateWorker(jobs <-chan int, rate int, done <-chan struct{}) <-chan int {
	// TODO: implement
	return nil
}

func main() {
	fmt.Println("== Exercise 1: Buffered collect ==")
	// TODO: call collectResults(5) and print each result

	fmt.Println("== Exercise 2: First wins ==")
	// TODO: create 3 source channels that send after different delays,
	// call firstOf, and print whether you got a result or timed out.

	fmt.Println("== Exercise 3: Parallel map ==")
	// TODO: call parallelMap([]int{1,2,3,4,5}, func(x int) int { return x*x })
	// and print the result. Expected: [1 4 9 16 25]

	fmt.Println("== Challenge: Rate-limited worker ==")
	// TODO: set up jobs channel, launch rateWorker, wait 300ms, close done,
	// drain output, print count.
	_ = sync.WaitGroup{} // keep import if needed
	_ = time.Second      // keep import if needed
}

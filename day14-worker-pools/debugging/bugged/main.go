// Day 14 debugging — BUGGED.
//
// Scenario: a bounded worker pool squares numbers. A producer feeds `jobs`,
// N workers read jobs and write to `results`, and the consumer ranges over
// `results` to collect them.
//
// Bug: NOBODY closes the `results` channel. The producer correctly closes
// `jobs`, so the workers' `range jobs` loops end and the workers return — but
// because results is never closed, the consumer's `for r := range results`
// blocks forever after draining the last result, waiting for a close that never
// comes. Classic worker-pool deadlock: the program hangs.
//
// To avoid hanging the grader forever, a watchdog goroutine fires after a short
// timeout, reports the deadlock with evidence, and exits non-zero.
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func square(n int) int { return n * n }

func main() {
	const workers = 4
	const numJobs = 20

	jobs := make(chan int)
	results := make(chan int)

	// Producer: feed jobs, then close jobs (this part is correct).
	go func() {
		for i := 1; i <= numJobs; i++ {
			jobs <- i
		}
		close(jobs) // workers' range jobs will end
	}()

	// Worker pool: read jobs, write results.
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs { // ends when jobs is closed
				results <- square(j)
			}
		}()
	}

	// BUG: we never close(results) after the workers finish.
	// The missing line is:
	//     go func() { wg.Wait(); close(results) }()

	// Watchdog: if the consumer hangs, report and bail so we don't hang forever.
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			// consumer finished normally (won't happen in the bugged version)
		case <-time.After(2 * time.Second):
			fmt.Println("DEADLOCK detected: consumer is blocked on `range results`")
			fmt.Println("  cause: results channel was never closed after wg.Wait()")
			fmt.Printf("  workers all returned (jobs closed), but range results never sees a close\n")
			os.Exit(1)
		}
	}()

	// Consumer: collect results.
	sum := 0
	count := 0
	for r := range results { // BUG: blocks forever after the last result
		sum += r
		count++
	}
	close(done)

	fmt.Printf("collected %d results, sum of squares = %d\n", count, sum)
}

// Day 15 debugging — BUGGED.
//
// Scenario: a pipeline gen -> square that streams numbers. The consumer only
// wants the FIRST few results, so it `break`s out of the range early.
//
// Bug: no stage propagates cancellation. When the consumer breaks early, the
// `square` stage is blocked on `out <- n*n` (nobody is receiving anymore), and
// the `gen` stage is blocked on its own send to square. Both upstream goroutines
// leak — they never exit. The consumer returning does NOT tear down the
// pipeline.
//
// We prove the leak with runtime.NumGoroutine(): it stays elevated after the
// consumer is done, instead of returning to baseline.
package main

import (
	"fmt"
	"runtime"
	"time"
)

// gen streams a large (effectively unbounded for our purposes) sequence.
func gen() <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; ; i++ {
			out <- i // BUG: no cancellation — blocks forever once downstream stops
		}
	}()
	return out
}

// square reads from in and emits squares.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n // BUG: no cancellation — blocks forever once consumer stops
		}
	}()
	return out
}

func main() {
	base := runtime.NumGoroutine()

	// Consume only the first 3 results, then stop caring.
	results := square(gen())
	taken := 0
	for v := range results {
		fmt.Println("got:", v)
		taken++
		if taken == 3 {
			break // consumer leaves early — upstream is now orphaned
		}
	}

	// Let the orphaned goroutines settle (they're blocked on sends).
	time.Sleep(100 * time.Millisecond)

	leaked := runtime.NumGoroutine() - base
	fmt.Printf("after taking %d results: %d goroutines leaked (gen + square stranded)\n", taken, leaked)
	if leaked > 0 {
		fmt.Println("GOROUTINE LEAK: gen blocked on `out <- i`, square blocked on `out <- n*n`")
		fmt.Println("  the consumer's early break never told upstream to stop")
	}
}

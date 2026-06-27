// Day 14 debugging — FIXED.
//
// Same bounded worker pool, but results is closed correctly.
//
// Fix: after all workers finish (wg.Wait()), close(results). Because the workers
// are the only senders on results, the right place to close is a single
// coordinator goroutine that waits for all of them, then closes once:
//
//     go func() { wg.Wait(); close(results) }()
//
// Now the consumer's `for r := range results` ends cleanly when the last result
// is drained. No hang, no leak, clean under -race.
package main

import (
	"fmt"
	"sync"
)

func square(n int) int { return n * n }

func main() {
	const workers = 4
	const numJobs = 20

	jobs := make(chan int)
	results := make(chan int)

	// Producer: feed jobs, then close jobs.
	go func() {
		for i := 1; i <= numJobs; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Worker pool: read jobs, write results.
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results <- square(j)
			}
		}()
	}

	// FIX: one coordinator closes results after ALL workers are done.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Consumer: collect results until results is closed & drained.
	sum := 0
	count := 0
	for r := range results {
		sum += r
		count++
	}

	fmt.Printf("collected %d results, sum of squares = %d\n", count, sum)
	if count == numJobs {
		fmt.Println("all jobs processed, results channel closed cleanly — no hang, no leak")
	}
}

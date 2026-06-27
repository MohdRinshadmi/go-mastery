// Day 12 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ---- Exercise 1 ----

func collectResults(n int) []string {
	ch := make(chan string, n) // buffered: goroutines never block
	for id := 0; id < n; id++ {
		id := id
		go func() {
			time.Sleep(time.Duration(id*5) * time.Millisecond)
			ch <- fmt.Sprintf("result-%d", id)
		}()
	}
	results := make([]string, 0, n)
	for i := 0; i < n; i++ {
		results = append(results, <-ch)
	}
	return results
}

// ---- Exercise 2 ----

func mergeStrings(sources []<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup
	for _, src := range sources {
		wg.Add(1)
		src := src
		go func() {
			defer wg.Done()
			for v := range src {
				merged <- v
			}
		}()
	}
	// Close merged when all sources are done.
	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

func firstOf(sources []<-chan string, timeout time.Duration) (string, bool) {
	merged := mergeStrings(sources)
	select {
	case v, ok := <-merged:
		if !ok {
			return "", false
		}
		return v, true
	case <-time.After(timeout):
		return "", false
	}
}

// ---- Exercise 3 ----

func parallelMap(inputs []int, fn func(int) int) []int {
	results := make([]int, len(inputs))
	var wg sync.WaitGroup
	for i, v := range inputs {
		wg.Add(1)
		i, v := i, v
		go func() {
			defer wg.Done()
			results[i] = fn(v) // safe: each goroutine writes unique index
		}()
	}
	wg.Wait()
	return results
}

// ---- Challenge: Rate-limited worker ----

func rateWorker(jobs <-chan int, rate int, done <-chan struct{}) <-chan int {
	out := make(chan int)
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	go func() {
		defer close(out)
		defer ticker.Stop()
		for {
			// First: acquire a rate-limit token (wait for tick).
			select {
			case <-done:
				return
			case <-ticker.C:
				// Got permission to process one job.
			}
			// Then: get a job.
			select {
			case <-done:
				return
			case j, ok := <-jobs:
				if !ok {
					return // jobs channel closed
				}
				out <- j * 2
			}
		}
	}()
	return out
}

func main() {
	fmt.Println("== Exercise 1: Buffered collect ==")
	results := collectResults(5)
	sort.Strings(results)
	for _, r := range results {
		fmt.Println(" ", r)
	}

	fmt.Println("== Exercise 2: First wins ==")
	// Three sources: one fast, one medium, one slow.
	makeDelayed := func(val string, delay time.Duration) <-chan string {
		ch := make(chan string, 1)
		go func() {
			time.Sleep(delay)
			ch <- val
			close(ch)
		}()
		return ch
	}
	sources := []<-chan string{
		makeDelayed("slow",   200*time.Millisecond),
		makeDelayed("fast",   10*time.Millisecond),
		makeDelayed("medium", 50*time.Millisecond),
	}
	if v, ok := firstOf(sources, 500*time.Millisecond); ok {
		fmt.Printf("  first result: %q\n", v)
	} else {
		fmt.Println("  timed out")
	}

	// Timeout scenario.
	slowSources := []<-chan string{
		makeDelayed("too-late", 1*time.Second),
	}
	if v, ok := firstOf(slowSources, 50*time.Millisecond); ok {
		fmt.Printf("  got: %q\n", v)
	} else {
		fmt.Println("  timed out (expected)")
	}

	fmt.Println("== Exercise 3: Parallel map ==")
	squared := parallelMap([]int{1, 2, 3, 4, 5}, func(x int) int { return x * x })
	fmt.Println(" ", squared)

	fmt.Println("== Challenge: Rate-limited worker ==")
	jobs := make(chan int, 10)
	for i := 1; i <= 10; i++ {
		jobs <- i
	}
	// Don't close jobs — rateWorker will drain it until done fires.

	done := make(chan struct{})
	out := rateWorker(jobs, 5, done)

	// Let it run for 300ms (expect ~1-2 jobs at 5/sec).
	time.Sleep(300 * time.Millisecond)
	close(done)

	count := 0
	for range out {
		count++
	}
	fmt.Printf("  processed %d jobs in 300ms at 5/sec rate\n", count)
}

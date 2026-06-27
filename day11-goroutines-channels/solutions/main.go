// Day 11 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"fmt"
	"sort"
)

// ---- Exercise 1 ----

func sayHello(name string, out chan<- string) {
	out <- fmt.Sprintf("Hello, %s!", name)
}

// ---- Exercise 2 ----

func produce(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * 2
		}
	}()
	return out
}

// ---- Exercise 3 ----

func drain(ch <-chan int) []int {
	var result []int
	for {
		v, ok := <-ch
		if !ok {
			return result
		}
		result = append(result, v)
	}
}

// ---- Challenge: Fan-out ----

// fanOut reads each value from in exactly once and broadcasts it to all n outputs.
// A coordinator goroutine owns the read from in; n forwarder goroutines are not
// needed here — the coordinator sends to all outputs directly.
func fanOut(in <-chan int, n int) []<-chan int {
	// Create n output channels and wrap them as receive-only for callers.
	outputs := make([]chan int, n)
	result := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		outputs[i] = make(chan int)
		result[i] = outputs[i]
	}

	// One coordinator reads from in and fans out to all outputs.
	go func() {
		// Close all outputs when in is exhausted.
		defer func() {
			for _, out := range outputs {
				close(out)
			}
		}()
		for v := range in {
			// Send the same value to every output channel.
			// NOTE: this is sequential per value — a real broadcast with
			// goroutine-per-output is shown in Day 15's fan-in/fan-out section.
			for _, out := range outputs {
				out <- v
			}
		}
	}()
	return result
}

func main() {
	fmt.Println("== Exercise 1: Concurrent greeting ==")
	ch := make(chan string, 3) // buffered so goroutines don't block if we're slow
	names := []string{"Alice", "Bob", "Carol"}
	for _, name := range names {
		go sayHello(name, ch)
	}
	greetings := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		greetings = append(greetings, <-ch)
	}
	sort.Strings(greetings) // sort for deterministic output
	for _, g := range greetings {
		fmt.Println(" ", g)
	}

	fmt.Println("== Exercise 2: Pipeline double ==")
	for v := range double(produce(1, 2, 3, 4, 5)) {
		fmt.Printf("  doubled: %d\n", v)
	}

	fmt.Println("== Exercise 3: Closed-channel drain ==")
	numCh := make(chan int, 4)
	numCh <- 10
	numCh <- 20
	numCh <- 30
	numCh <- 40
	close(numCh)
	vals := drain(numCh)
	fmt.Printf("  drained: %v\n", vals)

	fmt.Println("== Challenge: Fan-out ==")
	// Produce 1, 2, 3 and fan out to 3 channels.
	// Each value should appear in all 3 output channels.
	outs := fanOut(produce(1, 2, 3), 3)

	// IMPORTANT: We must drain all output channels concurrently.
	// The coordinator sends to outputs[0], then [1], then [2] for each value.
	// If we drain channel 0 fully first, the coordinator blocks on channel 1.
	// Solution: launch a goroutine per output channel and collect results.
	type result struct {
		idx  int
		vals []int
	}
	resultCh := make(chan result, len(outs))
	for i, out := range outs {
		i, out := i, out
		go func() {
			var vals []int
			for v := range out {
				vals = append(vals, v)
			}
			resultCh <- result{idx: i, vals: vals}
		}()
	}
	results := make([]result, len(outs))
	for range outs {
		r := <-resultCh
		results[r.idx] = r
	}
	for _, r := range results {
		fmt.Printf("  channel %d received: %v\n", r.idx, r.vals)
	}
}

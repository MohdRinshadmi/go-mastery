// Day 11 — Goroutines & Channels walkthrough
// Run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson.
// Change things and re-run to see what breaks — that's how you build intuition.
package main

import (
	"fmt"
	"time"
)

// ---- Section 1: simplest goroutine + channel rendezvous ----

// greet sends a greeting on the channel then returns.
// Note the send-only annotation: this function cannot receive.
func greet(name string, out chan<- string) {
	out <- fmt.Sprintf("Hello, %s!", name)
}

// ---- Section 2: generator — the canonical producer pattern ----

// generate feeds integers onto a channel from a goroutine and closes when done.
// Returns a receive-only channel: callers can only read, not close or send.
// This is the idiomatic "producer" signature you'll write constantly.
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // close when done — signals range loops to stop
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// ---- Section 3: pipeline stage — square ----

// square reads ints from in, squares them, sends to out.
// Both channels are directional — can't be misused.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in { // exits when in is closed and drained
			out <- n * n
		}
	}()
	return out
}

// ---- Section 4: closed channel semantics ----

func closedChannelDemo() {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	// Receive with comma-ok: ok==false means closed and drained.
	for {
		v, ok := <-ch
		if !ok {
			fmt.Println("  channel closed")
			break
		}
		fmt.Printf("  received %d\n", v)
	}
}

// ---- Section 5: the classic loop-variable closure bug + fix ----

func loopClosureBugAndFix() {
	fmt.Println("  [BUG version — may print repeated values]")
	done := make(chan struct{})
	// BAD: all goroutines capture the same &i.
	// By the time they run, the loop may have finished.
	// We collect them on done to actually see output (and the bug).
	bugResults := make(chan int, 5)
	for i := 0; i < 5; i++ {
		go func() {
			bugResults <- i // captures &i — undefined behavior
			done <- struct{}{}
		}()
	}
	for j := 0; j < 5; j++ {
		<-done
	}
	close(bugResults)
	for v := range bugResults {
		fmt.Printf("  bug goroutine saw i=%d\n", v)
	}

	fmt.Println("  [FIX version — each goroutine gets its own i]")
	fixResults := make(chan int, 5)
	for i := 0; i < 5; i++ {
		i := i // rebind: new variable per iteration, copied by value
		go func() {
			fixResults <- i
			done <- struct{}{}
		}()
	}
	for j := 0; j < 5; j++ {
		<-done
	}
	close(fixResults)
	for v := range fixResults {
		fmt.Printf("  fix goroutine saw i=%d\n", v)
	}
}

// ---- Section 6: goroutine lifetime — using a channel as a "done" signal ----

// worker does some work and signals completion on the done channel.
// This is the simplest goroutine lifetime management: send on done when finished.
func worker(id int, done chan<- int) {
	// Simulate variable work duration.
	time.Sleep(time.Duration(id*10) * time.Millisecond)
	done <- id
}

func main() {
	fmt.Println("=== Section 1: Goroutine + Channel rendezvous ===")
	ch := make(chan string)
	go greet("Go Engineer", ch)
	msg := <-ch // blocks until goroutine sends
	fmt.Println(" ", msg)

	fmt.Println("=== Section 2: Generator pattern ===")
	for n := range generate(1, 2, 3, 4, 5) {
		fmt.Printf("  generated: %d\n", n)
	}

	fmt.Println("=== Section 3: Pipeline (generate → square) ===")
	// Chain two stages: numbers flow through both.
	for sq := range square(generate(2, 3, 4, 5)) {
		fmt.Printf("  squared: %d\n", sq)
	}

	fmt.Println("=== Section 4: Closed channel semantics ===")
	closedChannelDemo()

	fmt.Println("=== Section 5: Loop closure bug + fix ===")
	loopClosureBugAndFix()

	fmt.Println("=== Section 6: Goroutine lifetime with done channel ===")
	done := make(chan int, 5)
	for id := 1; id <= 5; id++ {
		go worker(id, done)
	}
	// Wait for all workers — collect their IDs as they finish.
	for i := 0; i < 5; i++ {
		id := <-done
		fmt.Printf("  worker %d finished\n", id)
	}

	fmt.Println("=== All sections complete ===")
}

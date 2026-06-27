// Day 12 — Buffered Channels, select, WaitGroups & sync.Once
// Run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson.
package main

import (
	"fmt"
	"sync"
	"time"
)

// ---- Section 1: Buffered channel behaviour ----

func bufferedChannelDemo() {
	ch := make(chan int, 3)

	// These three sends don't block — buffer has room.
	ch <- 10
	ch <- 20
	ch <- 30
	fmt.Printf("  buffered: len=%d cap=%d\n", len(ch), cap(ch))

	// Drain without a goroutine — possible because data is already buffered.
	fmt.Printf("  received: %d %d %d\n", <-ch, <-ch, <-ch)

	// Buffer-of-1 pattern: goroutine always finishes, even if caller doesn't receive.
	result := make(chan string, 1) // if nobody reads, goroutine still exits
	go func() {
		result <- "work done" // won't block even if caller times out
	}()
	select {
	case v := <-result:
		fmt.Printf("  got result: %s\n", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  timed out (goroutine still exits cleanly)")
	}
}

// ---- Section 2: select — multiplexing ----

func selectDemo() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	go func() {
		time.Sleep(10 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch2 <- "two"
	}()

	// Wait for whichever fires first, twice.
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println("  from ch1:", msg)
		case msg := <-ch2:
			fmt.Println("  from ch2:", msg)
		}
	}
}

// ---- Section 3: select with default (non-blocking) ----

func nonBlockingDemo() {
	ch := make(chan int, 1)

	// Try to receive — nothing there, default fires.
	select {
	case v := <-ch:
		fmt.Println("  got:", v)
	default:
		fmt.Println("  nothing ready yet (default)")
	}

	ch <- 42

	// Now something is ready.
	select {
	case v := <-ch:
		fmt.Println("  got:", v)
	default:
		fmt.Println("  still nothing")
	}
}

// ---- Section 4: select with timeout ----

// slowOperation simulates a slow upstream call.
func slowOperation(delay time.Duration) <-chan string {
	ch := make(chan string, 1) // buffer 1: goroutine never hangs if we time out
	go func() {
		time.Sleep(delay)
		ch <- "result"
	}()
	return ch
}

func timeoutDemo() {
	// Fast path: operation finishes before timeout.
	select {
	case v := <-slowOperation(10 * time.Millisecond):
		fmt.Println("  fast op:", v)
	case <-time.After(200 * time.Millisecond):
		fmt.Println("  fast op: timed out")
	}

	// Slow path: operation takes longer than timeout.
	select {
	case v := <-slowOperation(500 * time.Millisecond):
		fmt.Println("  slow op:", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("  slow op: timed out (expected)")
	}
}

// ---- Section 5: nil channel disables a select case ----

// This is an advanced select trick: nil channels never fire.
// Use it to dynamically disable a case.
func nilChannelDemo() {
	ch1 := make(chan int, 1)
	ch1 <- 1

	var ch2 chan int // nil: permanently disabled

	// Only ch1 case is active — ch2 case is silently ignored.
	select {
	case v := <-ch1:
		fmt.Println("  ch1:", v)
	case v := <-ch2: // nil channel: never selected
		fmt.Println("  ch2:", v)
	}
	fmt.Println("  (ch2 was nil — its case was disabled)")
}

// ---- Section 6: sync.WaitGroup ----

func waitGroupDemo() {
	var wg sync.WaitGroup
	results := make([]string, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1) // BEFORE go — critical
		go func(id int) {
			defer wg.Done() // always defer
			time.Sleep(time.Duration(id*5) * time.Millisecond)
			results[id] = fmt.Sprintf("worker-%d done", id)
		}(i)
	}

	wg.Wait() // block until all 5 goroutines call Done
	for _, r := range results {
		fmt.Println(" ", r)
	}
}

// ---- Section 7: sync.Once ----

var (
	once      sync.Once
	expensive *expensiveResource
)

type expensiveResource struct {
	data string
}

func initExpensive() {
	fmt.Println("  [once.Do running — should print exactly once]")
	expensive = &expensiveResource{data: "loaded"}
}

func getExpensive() *expensiveResource {
	once.Do(initExpensive)
	return expensive
}

func onceDemo() {
	var wg sync.WaitGroup
	// 10 goroutines all call getExpensive concurrently.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := getExpensive()
			_ = r // use the resource
		}()
	}
	wg.Wait()
	fmt.Printf("  resource data: %q\n", expensive.data)
}

func main() {
	fmt.Println("=== Section 1: Buffered channel ===")
	bufferedChannelDemo()

	fmt.Println("=== Section 2: select multiplexing ===")
	selectDemo()

	fmt.Println("=== Section 3: Non-blocking select (default) ===")
	nonBlockingDemo()

	fmt.Println("=== Section 4: Timeout with time.After ===")
	timeoutDemo()

	fmt.Println("=== Section 5: Nil channel disables case ===")
	nilChannelDemo()

	fmt.Println("=== Section 6: WaitGroup ===")
	waitGroupDemo()

	fmt.Println("=== Section 7: sync.Once ===")
	onceDemo()

	fmt.Println("=== All sections complete ===")
}

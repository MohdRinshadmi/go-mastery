// Day 11 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// Mentor review: I will look for goroutine lifetime clarity, correct channel
// direction annotations, and whether your range loops are deadlock-safe.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Concurrent greeting
//
// Implement sayHello(name string, out chan<- string) that sends
// "Hello, <name>!" on out.
//
// In main: launch 3 goroutines (Alice, Bob, Carol), collect their
// greetings from a channel, and print each one.
// HINT: You need to receive exactly 3 times — no more, no less.
// =====================================================================

func sayHello(name string, out chan<- string) {
	// TODO: implement
}

// =====================================================================
// EXERCISE 2 (beginner) — Pipeline: double the numbers
//
// Implement:
//   produce(nums ...int) <-chan int   — sends nums onto a channel; closes when done
//   double(in <-chan int) <-chan int  — reads ints, sends each * 2; closes when done
//
// In main: pipe produce(1, 2, 3, 4, 5) through double and print each result.
// Expected output (in order): 2 4 6 8 10
// =====================================================================

func produce(nums ...int) <-chan int {
	// TODO: implement. Remember to close the channel!
	return nil
}

func double(in <-chan int) <-chan int {
	// TODO: implement. Use range over in.
	return nil
}

// =====================================================================
// EXERCISE 3 (beginner) — Closed-channel detection
//
// Implement drain(ch <-chan int) []int that reads ALL values from ch
// (which will be closed by the caller) and returns them as a slice.
// Do NOT use range. Use the comma-ok idiom instead: v, ok := <-ch
//
// In main: make a buffered channel (capacity 4), send 10, 20, 30, 40,
// close it, then call drain and print the result.
// =====================================================================

func drain(ch <-chan int) []int {
	// TODO: implement using comma-ok (not range)
	return nil
}

// =====================================================================
// CHALLENGE (intermediate) — Fan-out with done signal
//
// Implement fanOut(in <-chan int, n int) []<-chan int that:
//   - Returns n receive-only channels
//   - Launches n goroutines, each reading from `in` and forwarding to its channel
//   - Each value from `in` should go to ALL n channels (broadcast)
//   - All n channels must be closed when `in` is closed
//
// HINTS:
//   - Each output channel gets a goroutine that reads from `in` and forwards.
//   - But here's the tricky part: with n goroutines all reading from ONE channel,
//     each value goes to ONE goroutine, not all. For true broadcast, you need
//     to read once then forward to ALL outputs. Think about who reads `in`.
//   - Use a coordinator goroutine that reads `in` and sends to all n outputs.
//
// In main: produce(1, 2, 3) → fanOut(_, 3) → print each channel's output.
// Each of the 3 values should appear in ALL 3 output channels.
// =====================================================================

func fanOut(in <-chan int, n int) []<-chan int {
	// TODO: implement
	return nil
}

func main() {
	fmt.Println("== Exercise 1: Concurrent greeting ==")
	// TODO: launch 3 goroutines, collect and print their greetings

	fmt.Println("== Exercise 2: Pipeline double ==")
	// TODO: pipe produce(1,2,3,4,5) through double and print each result

	fmt.Println("== Exercise 3: Closed-channel drain ==")
	// TODO: create channel, send 10/20/30/40, close, drain, print

	fmt.Println("== Challenge: Fan-out ==")
	// TODO: fanOut(produce(1,2,3), 3) and print each channel's output
}

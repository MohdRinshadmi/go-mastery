// Day 01 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// I (your mentor) will review this like a production PR. Write clean code.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Temperature converter
// Write celsiusToFahrenheit. Formula: F = C*9/5 + 32
// Then print the conversion for 0, 37, and 100 degrees Celsius.
// =====================================================================

func celsiusToFahrenheit(c float64) float64 {
	// TODO: implement
	return 0
}

// =====================================================================
// EXERCISE 2 (beginner) — Safe divide with the (result, error) pattern
// Return an error when dividing by zero. Use errors.New (add the import).
// In main, call it with (20, 4) and (5, 0) and handle BOTH cases.
// =====================================================================

func safeDivide(a, b int) (int, error) {
	// TODO: implement. Hint: import "errors"
	return 0, nil
}

// =====================================================================
// EXERCISE 3 (beginner) — Variadic average
// average(nums ...float64) float64 returns the mean.
// EDGE CASE: average of zero numbers should return 0, not panic (no div by 0).
// =====================================================================

func average(nums ...float64) float64 {
	// TODO: implement
	return 0
}

// =====================================================================
// CHALLENGE (intermediate) — Retry with closure + constant + defer
//
// Implement withRetry(op func() error) error that:
//   - calls op() up to MaxAttempts times (define MaxAttempts as a const = 3)
//   - returns nil on the first success
//   - if all attempts fail, returns the LAST error, wrapped so the message
//     reads like: "operation failed after 3 attempts: <original error>"
//     (use fmt.Errorf with %w  — we'll go deep on %w in Day 3, try it now)
//   - use defer to print "  cleanup ran" exactly once when withRetry returns
//
// Then in main: create a flaky operation using a closure that fails the
// first 2 times and succeeds on the 3rd. Prove withRetry recovers.
// =====================================================================

const MaxAttempts = 3

func withRetry(op func() error) error {
	// TODO: implement
	return nil
}

func main() {
	fmt.Println("== Exercise 1: Temperature ==")
	// TODO: print conversions for 0, 37, 100

	fmt.Println("== Exercise 2: Safe Divide ==")
	// TODO: call safeDivide(20, 4) and safeDivide(5, 0), handle both

	fmt.Println("== Exercise 3: Average ==")
	// TODO: print average(1,2,3,4) and average() (empty case)

	fmt.Println("== Challenge: withRetry ==")
	// TODO: build a flaky closure and run withRetry on it
}

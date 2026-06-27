// Day 01 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"errors"
	"fmt"
)

func celsiusToFahrenheit(c float64) float64 {
	return c*9/5 + 32
}

func safeDivide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func average(nums ...float64) float64 {
	if len(nums) == 0 { // guard the edge case — no panic, no div-by-zero
		return 0
	}
	var total float64
	for _, n := range nums {
		total += n
	}
	return total / float64(len(nums))
}

const MaxAttempts = 3

func withRetry(op func() error) error {
	defer fmt.Println("  cleanup ran")

	var lastErr error
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		if err := op(); err != nil {
			lastErr = err
			fmt.Printf("  attempt %d failed: %v\n", attempt, err)
			continue
		}
		fmt.Printf("  attempt %d succeeded\n", attempt)
		return nil
	}
	// %w wraps lastErr so callers can errors.Is/As it (Day 3 topic).
	return fmt.Errorf("operation failed after %d attempts: %w", MaxAttempts, lastErr)
}

func main() {
	fmt.Println("== Exercise 1: Temperature ==")
	for _, c := range []float64{0, 37, 100} {
		fmt.Printf("  %.0f°C = %.1f°F\n", c, celsiusToFahrenheit(c))
	}

	fmt.Println("== Exercise 2: Safe Divide ==")
	if q, err := safeDivide(20, 4); err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Println("  20 / 4 =", q)
	}
	if _, err := safeDivide(5, 0); err != nil {
		fmt.Println("  error:", err)
	}

	fmt.Println("== Exercise 3: Average ==")
	fmt.Printf("  average(1,2,3,4) = %.2f\n", average(1, 2, 3, 4))
	fmt.Printf("  average()        = %.2f\n", average())

	fmt.Println("== Challenge: withRetry ==")
	// Flaky closure: fails the first 2 calls, succeeds on the 3rd.
	calls := 0
	flaky := func() error {
		calls++
		if calls < 3 {
			return fmt.Errorf("transient network error (call %d)", calls)
		}
		return nil
	}
	if err := withRetry(flaky); err != nil {
		fmt.Println("  final:", err)
	} else {
		fmt.Println("  final: success ✅")
	}
}

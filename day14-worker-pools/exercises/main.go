// Day 14 — YOUR exercises. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
)

// =====================================================================
// EXERCISE 1 — bounded worker pool
// squarePool(nums []int, workers int) []int : N workers square numbers
// via a jobs channel. Remember: producer closes jobs; close results after
// a WaitGroup. Drain results into the return slice.
// =====================================================================

func squarePool(nums []int, workers int) []int {
	// TODO
	return nil
}

// =====================================================================
// EXERCISE 2 — fan-in
// merge(chans ...<-chan int) <-chan int : merge several channels into one,
// closing the output once all inputs are drained.
// =====================================================================

func merge(chans ...<-chan int) <-chan int {
	// TODO
	return nil
}

// =====================================================================
// CHALLENGE — concurrent checker with errgroup
// checkAll(ctx, items []int) (map[int]bool, error): for each item run a
// fake check concurrently with BOUNDED concurrency (SetLimit), collect
// results into a map (guard with a mutex), cancel remaining on first error.
// Use golang.org/x/sync/errgroup (go get golang.org/x/sync && go mod tidy).
// =====================================================================

func checkAll(ctx context.Context, items []int) (map[int]bool, error) {
	// TODO
	return nil, nil
}

func main() {
	fmt.Println("== Exercise 1 ==")
	// TODO: squarePool([]int{1..8}, 4)

	fmt.Println("== Exercise 2 ==")
	// TODO: merge a few channels

	fmt.Println("== Challenge ==")
	// TODO: checkAll
	_ = context.Background()
}

// Day 15 Phase-3 capstone — YOUR exercises. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"time"
)

// =====================================================================
// PART 1 — Concurrent URL checker (offline-safe)
// checkURLs(ctx, urls, workers) map[string]string : check each "url" with
// BOUNDED concurrency (worker pool) and a per-check timeout via context.
// Use the provided fakeCheck as the "request". Return url -> "ok"/"timeout".
// =====================================================================

// fakeCheck simulates a request that takes `latency`, respecting ctx.
func fakeCheck(ctx context.Context, url string, latency time.Duration) (string, error) {
	select {
	case <-time.After(latency):
		return "ok", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func checkURLs(ctx context.Context, urls map[string]time.Duration, workers int) map[string]string {
	// TODO: worker pool over urls; per-check context.WithTimeout; collect results
	return nil
}

// =====================================================================
// PART 2 — Cancellable pipeline
// Implement gen -> square -> filterEven, each stage selecting on ctx.Done().
// In main, consume only the first 2 results then cancel.
// =====================================================================

func gen(ctx context.Context, nums ...int) <-chan int {
	// TODO
	return nil
}
func square(ctx context.Context, in <-chan int) <-chan int {
	// TODO
	return nil
}
func filterEven(ctx context.Context, in <-chan int) <-chan int {
	// TODO
	return nil
}

func main() {
	fmt.Println("== Part 1: URL checker ==")
	// TODO: build a urls map with varied latencies, call checkURLs with a
	// timeout shorter than some latencies so a couple time out.
	_ = context.Background()

	fmt.Println("== Part 2: pipeline ==")
	// TODO
}

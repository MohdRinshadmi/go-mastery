// Day 22 debugging — the test that's green on your laptop and red in CI.
//
// summarize() builds a "k=v" summary of a map of feature flags. The unit
// test (main_test.go) asserts an exact string. It passes most of the time
// locally, then fails intermittently in CI — a classic FLAKY TEST that
// erodes trust in the pipeline.
//
// The root cause: Go RANDOMIZES map iteration order on purpose. Any code
// (or test) that assumes a fixed order is a latent bug.
//
// STDLIB ONLY. `go run .` here proves the non-determinism directly so you
// can see it without waiting for CI to flake.
package main

import (
	"fmt"
	"strings"
)

// summarize returns a "k=v,k=v,..." summary of flags.
//
// BUG: it ranges the map directly, so the order of the pairs is random on
// every run. Two correct-looking runs can produce different strings.
func summarize(flags map[string]string) string {
	parts := make([]string, 0, len(flags))
	for k, v := range flags { // random iteration order!
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}

func main() {
	flags := map[string]string{
		"beta":   "on",
		"cache":  "off",
		"region": "eu",
	}

	// Run it many times; collect the distinct outputs to expose the flake.
	seen := map[string]int{}
	for i := 0; i < 1000; i++ {
		seen[summarize(flags)]++
	}

	fmt.Printf("summarize() produced %d DISTINCT outputs across 1000 runs:\n", len(seen))
	for s, n := range seen {
		fmt.Printf("  %-28s x%d\n", s, n)
	}
	if len(seen) > 1 {
		fmt.Println("=> output is NON-DETERMINISTIC; any exact-match test is flaky")
	}
}

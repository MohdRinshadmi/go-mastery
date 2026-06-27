// Day 22 debugging — FIXED.
//
// The fix is to make summarize() DETERMINISTIC by sorting the keys before
// joining. A deterministic function has a stable output, so the exact-match
// test is no longer flaky. The lesson: never depend on map iteration order;
// sort keys whenever order is observable (output, hashing, golden files).
//
// STDLIB ONLY. `go run .` proves the output is now identical every time.
package main

import (
	"fmt"
	"sort"
	"strings"
)

// summarize returns a "k=v,k=v,..." summary of flags in STABLE key order.
func summarize(flags map[string]string) string {
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys) // FIX: impose a deterministic order

	parts := make([]string, 0, len(flags))
	for _, k := range keys {
		parts = append(parts, k+"="+flags[k])
	}
	return strings.Join(parts, ",")
}

func main() {
	flags := map[string]string{
		"beta":   "on",
		"cache":  "off",
		"region": "eu",
	}

	seen := map[string]int{}
	for i := 0; i < 1000; i++ {
		seen[summarize(flags)]++
	}

	fmt.Printf("summarize() produced %d DISTINCT output(s) across 1000 runs:\n", len(seen))
	for s, n := range seen {
		fmt.Printf("  %-28s x%d\n", s, n)
	}
	if len(seen) == 1 {
		fmt.Println("=> output is DETERMINISTIC; the exact-match test is now stable")
	}
}

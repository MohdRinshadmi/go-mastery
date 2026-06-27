package main

import (
	"fmt"
	"strings"
)

// parseFields splits a log line into fields. To "be efficient" it pre-sizes the
// slice with some spare capacity — which is exactly what plants the bug.
func parseFields(line string) []string {
	parts := strings.Fields(line)
	fields := make([]string, 0, len(parts)+4) // spare capacity: the trap
	for _, p := range parts {
		fields = append(fields, p)
	}
	return fields
}

// enrich returns the raw row and an enriched row (raw + a status column).
// The raw row must stay untouched.
func enrich(line, status string) (raw, enriched []string) {
	fields := parseFields(line)

	raw = fields // we want to keep the original around, untouched

	// BUG: `fields` still has spare capacity, so this append writes the status
	// IN PLACE into the shared backing array. `raw` aliases that same array.
	enriched = append(fields, status)

	return raw, enriched
}

func main() {
	line := "2026-06-27 login user=42"

	raw, enriched := enrich(line, "OK")

	// Simulate the raw row being used later (e.g. re-parsed or re-appended).
	raw = append(raw[:3], "EXTRA")

	fmt.Println("raw:     ", raw)
	fmt.Println("enriched:", enriched) // expected: [2026-06-27 login user=42 OK]
}

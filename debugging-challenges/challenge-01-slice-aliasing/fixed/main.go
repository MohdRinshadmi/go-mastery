package main

import (
	"fmt"
	"strings"
)

func parseFields(line string) []string {
	parts := strings.Fields(line)
	fields := make([]string, 0, len(parts)+4)
	for _, p := range parts {
		fields = append(fields, p)
	}
	return fields
}

func enrich(line, status string) (raw, enriched []string) {
	fields := parseFields(line)

	raw = fields

	// FIX: build a genuinely independent slice before appending, so the append
	// can never write into the array `raw` points at.
	enriched = make([]string, len(fields), len(fields)+1)
	copy(enriched, fields)
	enriched = append(enriched, status)

	// Equivalent one-liner using the three-index slice expression (cap == len
	// forces append to reallocate):
	//   enriched = append(fields[:len(fields):len(fields)], status)

	return raw, enriched
}

func main() {
	line := "2026-06-27 login user=42"

	raw, enriched := enrich(line, "OK")

	raw = append(raw[:3], "EXTRA")

	fmt.Println("raw:     ", raw)
	fmt.Println("enriched:", enriched)
}

package main

import "fmt"

// Get threads the map's comma-ok result through the generic signature.
//
// THE FIX: return (T, bool). The bool is the map's own "present" flag,
// so the caller can distinguish "key absent" from "key present with the
// zero value of T". The zero value of T alone is ambiguous; the bool
// removes the ambiguity.
func Get[T any](m map[string]T, key string) (T, bool) {
	v, ok := m[key]
	return v, ok
}

func main() {
	// Same data: "Dana" actually played and scored 0 points.
	scores := map[string]int{
		"Alice": 42,
		"Bob":   17,
		"Dana":  0, // Dana played but scored zero — a legitimate entry.
	}

	fmt.Println("=== FIXED: comma-ok generic lookup ===")

	report("Alice", scores)
	report("Dana", scores) // present with value 0 — reported correctly.
	report("Zoe", scores)  // absent — reported as not found.
}

// report uses the comma-ok bool to tell present-zero from missing.
func report(name string, scores map[string]int) {
	score, ok := Get(scores, name)
	if !ok {
		fmt.Printf("Player %s not found (never played).\n", name)
		return
	}
	fmt.Printf("Player %s scored %d points.\n", name, score)
}

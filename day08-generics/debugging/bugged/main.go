package main

import "fmt"

// Get is a generic map lookup helper. Looks clean, compiles fine.
//
// THE BUG: it returns a bare T. When the key is missing, map indexing
// hands back the zero value of T (0 for int, "" for string, nil for
// pointers). The caller has NO WAY to tell "the key was absent" from
// "the key was present and its value happens to be the zero value".
//
// For a scoreboard, the zero value 0 is a perfectly real score — a player
// who genuinely scored 0 looks identical to a player who was never added.
func Get[T any](m map[string]T, key string) T {
	return m[key]
}

func main() {
	// Real game data. Note: "Dana" actually played and scored 0 points.
	scores := map[string]int{
		"Alice": 42,
		"Bob":   17,
		"Dana":  0, // Dana played but scored zero — a legitimate entry.
	}

	fmt.Println("=== BUGGED: bare-zero generic lookup ===")

	// Look up a player who exists with a real non-zero score.
	report("Alice", Get(scores, "Alice"))

	// Look up Dana, who exists with a legitimate score of 0.
	report("Dana", Get(scores, "Dana"))

	// Look up "Zoe", who was NEVER added to the map.
	// The map hands back 0, and Get() can't signal "missing".
	report("Zoe", Get(scores, "Zoe"))

	fmt.Println()
	fmt.Println("Notice: Zoe is reported as 'scored 0 points' even though")
	fmt.Println("she never played. She is indistinguishable from Dana.")
}

// report prints a player's score. It trusts whatever Get returned —
// it has no signal that a player might be missing.
func report(name string, score int) {
	fmt.Printf("Player %s scored %d points.\n", name, score)
}

// Day 02 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// I (your mentor) will review this like a production PR. Write clean code.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Slice de-duplication
//
// Write unique(s []string) []string that returns a new slice containing
// only the first occurrence of each element (preserve insertion order).
//
// Example: unique(["a","b","a","c","b"]) → ["a","b","c"]
//
// Hint: use a map[string]bool (or map[string]struct{}) as a "seen" set.
// =====================================================================

func unique(s []string) []string {
	// TODO: implement
	return nil
}

// =====================================================================
// EXERCISE 2 (beginner) — Word frequency counter
//
// Write wordFreq(words []string) map[string]int that counts how many
// times each word appears.
//
// Example: wordFreq(["go","is","go"]) → map["go":2 "is":1]
//
// Then in main: print the results in ALPHABETICAL key order.
// Hint: sort.Strings(keys) after extracting keys.
// =====================================================================

func wordFreq(words []string) map[string]int {
	// TODO: implement
	return nil
}

// =====================================================================
// EXERCISE 3 (beginner) — Safe map lookup with default
//
// Write getOrDefault(m map[string]int, key string, def int) int:
// - returns m[key] if key exists
// - returns def if key does NOT exist
// - you MUST use the comma-ok idiom (not just m[key] with a zero check)
//
// Test with: m = {"a": 0}, key "a" must return 0 not the default
//            key "z" must return the default
// (This shows why comma-ok matters when zero IS a valid value.)
// =====================================================================

func getOrDefault(m map[string]int, key string, def int) int {
	// TODO: implement using comma-ok
	return def
}

// =====================================================================
// CHALLENGE (intermediate) — Safe slice operations
//
// Implement a mini "stack" using a slice as the backing store.
// The stack must:
//   1. Push(v int)          — append to end
//   2. Pop() (int, bool)    — remove and return last element;
//                             return (0, false) if empty (no panic!)
//   3. Peek() (int, bool)   — return last element without removing;
//                             return (0, false) if empty
//
// IMPORTANT: Push must not leak the aliasing gotcha.
// When you Pop, reslice the backing slice — do NOT leave old values
// accessible via cap. (Hint: think about what happens if you only do
// s = s[:len(s)-1] — the old value is still in the underlying array.)
//
// Use this struct:
//   type Stack struct { data []int }
// (NOT a pointer receiver — that's Day 3 material; but think about why
//  you'd need one here and add a TODO comment about it.)
// =====================================================================

type Stack struct {
	data []int
}

func (s Stack) Push(v int) Stack {
	// TODO: implement — return the new Stack (value receiver limitation)
	return s
}

func (s Stack) Pop() (Stack, int, bool) {
	// TODO: implement — return (newStack, value, found)
	return s, 0, false
}

func (s Stack) Peek() (int, bool) {
	// TODO: implement
	return 0, false
}

func main() {
	fmt.Println("== Exercise 1: unique ==")
	input := []string{"go", "is", "great", "go", "is", "fast"}
	result := unique(input)
	fmt.Println(" ", result) // [go is great fast]

	fmt.Println("== Exercise 2: wordFreq (sorted output) ==")
	words := []string{"the", "quick", "brown", "fox", "the", "quick", "the"}
	freq := wordFreq(words)
	// TODO: extract keys, sort them, print in sorted order

	fmt.Println("== Exercise 3: getOrDefault ==")
	m := map[string]int{"a": 0, "b": 42}
	fmt.Println("  a:", getOrDefault(m, "a", -1)) // 0  (not -1!)
	fmt.Println("  b:", getOrDefault(m, "b", -1)) // 42
	fmt.Println("  z:", getOrDefault(m, "z", -1)) // -1

	fmt.Println("== Challenge: Stack ==")
	var st Stack
	st = st.Push(10)
	st = st.Push(20)
	st = st.Push(30)

	if v, ok := st.Peek(); ok {
		fmt.Println("  peek:", v) // 30
	}
	var val int
	var ok bool
	st, val, ok = st.Pop()
	fmt.Printf("  pop: val=%d ok=%t\n", val, ok) // 30 true
	st, val, ok = st.Pop()
	fmt.Printf("  pop: val=%d ok=%t\n", val, ok) // 20 true
	st, val, ok = st.Pop()
	fmt.Printf("  pop: val=%d ok=%t\n", val, ok) // 10 true
	st, val, ok = st.Pop()
	fmt.Printf("  pop empty: val=%d ok=%t\n", val, ok) // 0 false
	_ = freq
}

// Day 02 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"fmt"
	"sort"
)

func unique(s []string) []string {
	seen := make(map[string]struct{}, len(s)) // pre-size: at most len(s) unique entries
	result := make([]string, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func wordFreq(words []string) map[string]int {
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++ // zero-value of int is 0, so first increment goes 0→1 safely
	}
	return freq
}

func getOrDefault(m map[string]int, key string, def int) int {
	// comma-ok is REQUIRED: if the value can be 0, checking v == 0 is wrong.
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

// Stack backed by a slice. Using value receivers here forces callers to
// assign the returned Stack — that makes mutation explicit.
//
// TODO (Day 3 preview): a pointer receiver (*Stack) would let Push/Pop mutate
// in-place without returning a new value. Most real stacks use *Stack.
// For now, value receiver teaches the aliasing concepts without pointer confusion.
type Stack struct {
	data []int
}

func (s Stack) Push(v int) Stack {
	// append may allocate a new backing array — that's fine.
	// No aliasing risk here since we return a new Stack header each time.
	s.data = append(s.data, v)
	return s
}

func (s Stack) Pop() (Stack, int, bool) {
	if len(s.data) == 0 {
		return s, 0, false
	}
	n := len(s.data)
	val := s.data[n-1]
	// Zero out the slot before reslicing so the old value doesn't remain
	// accessible through the underlying array via cap. This prevents subtle
	// memory leaks when the element is a pointer or large struct.
	s.data[n-1] = 0
	s.data = s.data[:n-1]
	return s, val, true
}

func (s Stack) Peek() (int, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	return s.data[len(s.data)-1], true
}

func main() {
	fmt.Println("== Exercise 1: unique ==")
	input := []string{"go", "is", "great", "go", "is", "fast"}
	fmt.Println(" ", unique(input)) // [go is great fast]

	fmt.Println("== Exercise 2: wordFreq (sorted output) ==")
	words := []string{"the", "quick", "brown", "fox", "the", "quick", "the"}
	freq := wordFreq(words)
	keys := make([]string, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, freq[k])
	}

	fmt.Println("== Exercise 3: getOrDefault ==")
	m := map[string]int{"a": 0, "b": 42}
	fmt.Println("  a:", getOrDefault(m, "a", -1)) // 0  (not -1! comma-ok matters)
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
}

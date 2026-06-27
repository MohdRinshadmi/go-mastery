// Day 08 walkthrough — generics. Run: go run main.go
package main

import (
	"cmp"   // Go 1.21+: Ordered constraint + helpers
	"fmt"
)

// ---- 1. Generic functions ----------------------------------------------

// Map transforms []T -> []U. Two type params; both inferred at call site.
func Map[T, U any](s []T, fn func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}

func Filter[T any](s []T, keep func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// ---- 2. Constraints ----------------------------------------------------

// cmp.Ordered constrains to types supporting < <= >= > (ints, floats, strings).
func Max[T cmp.Ordered](s []T) (T, bool) {
	var zero T
	if len(s) == 0 {
		return zero, false
	}
	m := s[0]
	for _, v := range s[1:] {
		if v > m {
			m = v
		}
	}
	return m, true
}

// Custom constraint via interface union — only number types.
type Number interface {
	~int | ~int64 | ~float64 // ~ means "any type whose underlying type is this"
}

func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// ---- 3. Generic type ---------------------------------------------------

// Stack[T] — a type-safe stack for any element type.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}

func main() {
	fmt.Println("== Map / Filter ==")
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("  doubled:", Map(nums, func(x int) int { return x * 2 }))
	fmt.Println("  evens:  ", Filter(nums, func(x int) bool { return x%2 == 0 }))
	fmt.Println("  lengths:", Map([]string{"go", "rust"}, func(s string) int { return len(s) }))

	fmt.Println("== Constraints ==")
	m, _ := Max([]int{3, 9, 2})
	fmt.Println("  max int:", m)
	ms, _ := Max([]string{"banana", "apple", "cherry"})
	fmt.Println("  max str:", ms)
	fmt.Println("  sum:", Sum([]float64{1.5, 2.5, 3.0}))

	fmt.Println("== Generic Stack[T] ==")
	var st Stack[string]
	st.Push("a")
	st.Push("b")
	v, _ := st.Pop()
	fmt.Println("  popped:", v)
}

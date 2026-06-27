// Day 08 — YOUR exercises. Run: go run main.go
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Generic Contains
// Contains[T comparable](s []T, target T) bool — true if target is in s.
// Why "comparable" not "any"? Because we use ==. Try changing it to any
// and read the compiler error — that's the lesson.
// =====================================================================

func Contains[T comparable](s []T, target T) bool {
	// TODO
	return false
}

// =====================================================================
// EXERCISE 2 (beginner) — Generic Keys
// Keys[K comparable, V any](m map[K]V) []K returns all keys (any order).
// =====================================================================

func Keys[K comparable, V any](m map[K]V) []K {
	// TODO
	return nil
}

// =====================================================================
// EXERCISE 3 (beginner) — Reduce
// Reduce[T, A any](s []T, init A, fn func(A, T) A) A folds the slice.
// e.g. Reduce([]int{1,2,3}, 0, func(acc, x int) int { return acc + x }) == 6
// =====================================================================

func Reduce[T, A any](s []T, init A, fn func(A, T) A) A {
	// TODO
	return init
}

// =====================================================================
// CHALLENGE (intermediate) — generic OrderedSet[T]
// Build type OrderedSet[T comparable] backed by a map[T]struct{} plus a
// slice to preserve insertion order. Methods:
//   Add(v T)            // no-op if already present
//   Has(v T) bool
//   Items() []T         // in insertion order
// Prove it works with both int and string element types in main.
// =====================================================================

type OrderedSet[T comparable] struct {
	// TODO: fields
}

func main() {
	fmt.Println("== Exercise 1 ==")
	// TODO: Contains([]int{1,2,3}, 2)

	fmt.Println("== Exercise 2 ==")
	// TODO: Keys(map[string]int{"a":1,"b":2})

	fmt.Println("== Exercise 3 ==")
	// TODO: Reduce sum

	fmt.Println("== Challenge ==")
	// TODO: OrderedSet with ints and strings
}

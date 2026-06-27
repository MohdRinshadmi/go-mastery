// Day 08 — reference solutions. Run: go run main.go
package main

import "fmt"

func Contains[T comparable](s []T, target T) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func Reduce[T, A any](s []T, init A, fn func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = fn(acc, v)
	}
	return acc
}

// Challenge: insertion-ordered set
type OrderedSet[T comparable] struct {
	seen  map[T]struct{}
	order []T
}

func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{seen: make(map[T]struct{})}
}

func (s *OrderedSet[T]) Add(v T) {
	if _, ok := s.seen[v]; ok {
		return
	}
	s.seen[v] = struct{}{}
	s.order = append(s.order, v)
}

func (s *OrderedSet[T]) Has(v T) bool {
	_, ok := s.seen[v]
	return ok
}

func (s *OrderedSet[T]) Items() []T {
	out := make([]T, len(s.order))
	copy(out, s.order) // defensive copy — don't leak internal slice
	return out
}

func main() {
	fmt.Println("== Exercise 1 ==")
	fmt.Println("  Contains([1,2,3], 2):", Contains([]int{1, 2, 3}, 2))
	fmt.Println("  Contains([a,b], z):  ", Contains([]string{"a", "b"}, "z"))

	fmt.Println("== Exercise 2 ==")
	fmt.Println("  Keys:", Keys(map[string]int{"a": 1, "b": 2}))

	fmt.Println("== Exercise 3 ==")
	sum := Reduce([]int{1, 2, 3}, 0, func(acc, x int) int { return acc + x })
	fmt.Println("  sum:", sum)

	fmt.Println("== Challenge ==")
	si := NewOrderedSet[int]()
	for _, v := range []int{3, 1, 3, 2, 1} {
		si.Add(v)
	}
	fmt.Println("  int set items:", si.Items(), "has 2:", si.Has(2))

	ss := NewOrderedSet[string]()
	ss.Add("go")
	ss.Add("go")
	ss.Add("rust")
	fmt.Println("  str set items:", ss.Items())
}

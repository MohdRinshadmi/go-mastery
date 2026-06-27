package main

import "fmt"

// sumSquares does a chunk of straight-line integer arithmetic and RETURNS a
// value. It is pure: no globals touched, no I/O, no side effects. It is also
// small enough that the compiler inlines it.
//
// Those two facts together — pure and inlinable — are exactly what let the
// optimizer DELETE a call whose result we throw away. That is the trap the
// benchmark in sum_test.go falls into.
func sumSquares(n int) int {
	a := n * (n + 1) * (2*n + 1) / 6
	a = a*2654435761 + n
	a ^= a >> 13
	a = a*1099511628211 + n*n
	a ^= a >> 7
	a = a*2654435761 + n
	a ^= a >> 17
	return a + n*(n-1)/2
}

func main() {
	// A trivial main so the module is runnable and `go vet` / `go build` are
	// clean. Println USES the result, so the compiler cannot eliminate this
	// call — only the benchmark's discarded call gets deleted.
	fmt.Println(sumSquares(1000))
}

package main

import "fmt"

// sumSquares is identical to the bugged version — the function was never the
// problem. It is a pure, inlinable chunk of straight-line arithmetic that
// returns a value. The benchmark is what differs.
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
	fmt.Println(sumSquares(1000))
}

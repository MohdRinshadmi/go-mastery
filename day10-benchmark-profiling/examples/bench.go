// Day 10 examples — code we benchmark. Run: go test -bench=. -benchmem ./...
package bench

// Naive: lets append reallocate repeatedly as it grows.
func BuildNaive(n int) []int {
	var s []int // nil, cap 0
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// Prealloc: one allocation, no regrows.
func BuildPrealloc(n int) []int {
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// String building: the += anti-pattern vs strings.Builder is in the test.

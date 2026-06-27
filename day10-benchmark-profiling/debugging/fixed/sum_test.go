package main

import "testing"

// arg is a package-level variable so sumSquares(arg) is not constant-folded.
var arg = 1000

// sink is a package-level variable. Assigning the benchmark's result to it
// makes the work OBSERVABLE outside the loop: the compiler can no longer prove
// the call is dead, so it must actually run sumSquares every iteration. This
// is the classic, portable fix for dead-code elimination in benchmarks.
var sink int

// FIXED BENCHMARK.
//
// Same classic `for i := 0; i < b.N; i++` loop as the bugged version, but the
// result of sumSquares is stored into the package-level `sink`. Now the timing
// is honest: a realistic ns/op that grows if the function does more work.
//
// Modern alternative (Go 1.24+): `for b.Loop() { sink = sumSquares(arg) }`.
// b.Loop is designed so the compiler keeps the loop body alive, so it also
// prevents this elimination without needing the sink trick.
func BenchmarkSumSquares(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = sumSquares(arg)
	}
	sink = r // publish the result so the work cannot be optimized away
}

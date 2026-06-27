package main

import "testing"

// arg is a package-level variable, so sumSquares(arg) is NOT a compile-time
// constant the optimizer could just fold to a literal. The only reason the
// call disappears in the bugged benchmark below is that its result is unused.
var arg = 1000

// BUGGED BENCHMARK.
//
// We call sumSquares(arg) inside the b.N loop but never USE the result.
// sumSquares is pure and inlinable, so after inlining the compiler sees a
// block of arithmetic whose output nothing reads — dead code — and deletes it
// entirely ("dead-code elimination").
//
// The benchmark then times an essentially empty loop. You get an absurd
// ~0.3 ns/op that does NOT grow even if sumSquares did ten times more work.
// That number is a lie: it is not measuring sumSquares at all.
//
// We use the CLASSIC `for i := 0; i < b.N; i++` loop on purpose. The modern
// `for b.Loop()` form is built to DEFEAT this elimination, so it would hide
// the bug. See ../fixed for the honest version.
func BenchmarkSumSquares(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sumSquares(arg) // result discarded -> compiler deletes the call
	}
}

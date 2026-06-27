package main

import "testing"

// Run: go test -bench=. -benchmem
// The pooled Render shows far fewer allocs/op than bugged/ (the per-call buffer
// allocation is gone; the remaining alloc is the result copy the caller needs).
func BenchmarkRender(b *testing.B) {
	r := Record{ID: 42, Name: "widget", Score: 7}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Render(r)
	}
}

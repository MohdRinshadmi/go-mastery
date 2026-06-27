package main

import "testing"

// Run: go test -bench=. -benchmem
// Watch the "allocs/op" column — the bugged Render allocates a buffer per call.
func BenchmarkRender(b *testing.B) {
	r := Record{ID: 42, Name: "widget", Score: 7}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Render(r)
	}
}

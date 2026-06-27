// Day 29 debugging — allocation in a hot path (a bytes.Buffer per call).
//
// Render() formats a record into bytes. It's called on a very hot path (think:
// serializing every job/event/log line). Each call allocates a fresh
// bytes.Buffer that lives only for the call — exactly the short-lived, frequently
// allocated object that sync.Pool exists for. The allocations pile up as GC work
// and dominate the profile under load.
//
// Run the benchmark to see the allocations:
//   go test -bench=. -benchmem
// Run the program for a quick allocs-per-op estimate:
//   go run .
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
)

type Record struct {
	ID    int
	Name  string
	Score int
}

// Render formats one record. BUG: allocates a new buffer every call.
func Render(r Record) []byte {
	var buf bytes.Buffer // fresh heap allocation on every single call
	buf.WriteString("id=")
	buf.WriteString(strconv.Itoa(r.ID))
	buf.WriteString(" name=")
	buf.WriteString(r.Name)
	buf.WriteString(" score=")
	buf.WriteString(strconv.Itoa(r.Score))
	return append([]byte(nil), buf.Bytes()...) // copy out so caller owns it
}

func main() {
	const N = 200_000
	var m0, m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m0)

	var sink int
	for i := 0; i < N; i++ {
		out := Render(Record{ID: i, Name: "widget", Score: i % 100})
		sink += len(out)
	}

	runtime.ReadMemStats(&m1)
	allocs := m1.Mallocs - m0.Mallocs
	fmt.Printf("rendered %d records, sink=%d\n", N, sink)
	fmt.Printf("heap allocations: %d (~%.1f allocs/record)\n",
		allocs, float64(allocs)/float64(N))
	fmt.Println("BUG: a buffer is allocated on every call — pool it on the hot path.")
}

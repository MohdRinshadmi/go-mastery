// Day 29 debugging — FIXED: reuse buffers via sync.Pool on the hot path.
//
// The buffer is short-lived and allocated millions of times — the textbook case
// for sync.Pool. We Get a buffer from the pool, MUST Reset it (or we'd leak the
// previous call's bytes), use it, copy the result out for the caller, then Put it
// back for reuse. The per-call buffer allocation disappears.
//
//   go test -bench=. -benchmem   # compare allocs/op against bugged/
//   go run .
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

type Record struct {
	ID    int
	Name  string
	Score int
}

// bufPool reuses bytes.Buffers across calls instead of allocating each time.
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// Render formats one record, reusing a pooled buffer.
func Render(r Record) []byte {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()         // CRUCIAL: clear leftover bytes from a previous use
	defer bufPool.Put(buf)

	buf.WriteString("id=")
	buf.WriteString(strconv.Itoa(r.ID))
	buf.WriteString(" name=")
	buf.WriteString(r.Name)
	buf.WriteString(" score=")
	buf.WriteString(strconv.Itoa(r.Score))

	// Copy out so the caller owns the bytes; the buffer goes back to the pool.
	return append([]byte(nil), buf.Bytes()...)
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
	fmt.Println("OK: buffers are reused from a sync.Pool — far fewer allocations.")
}

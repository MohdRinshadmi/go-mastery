# Day 10 Notes — Benchmarking & Profiling Cheatsheet

Quick reference. Standard library + `golang.org/x/perf` (benchstat) only.

---

## Benchmark skeleton — two forms

```go
package mypkg

import "testing"

// sink defeats dead-code elimination (classic-loop form).
var sink int

// Classic form: framework scales b.N until timing is stable.
func BenchmarkWorkClassic(b *testing.B) {
    setup()           // one-time, expensive
    b.ResetTimer()    // <-- don't count setup
    var r int
    for i := 0; i < b.N; i++ {
        r = work()
    }
    sink = r          // <-- publish result so it isn't optimized away
}

// Modern form (Go 1.24+): prevents dead-code elimination automatically,
// excludes setup before the loop, no ResetTimer/sink needed.
func BenchmarkWorkLoop(b *testing.B) {
    setup()
    for b.Loop() {
        sink = work()
    }
}
```

---

## Running benchmarks

```bash
go test -bench=.            -run=^$              # all benchmarks, skip tests
go test -bench=BenchmarkX   -run=^$              # one (regexp match)
go test -bench=. -benchmem  -run=^$              # add B/op and allocs/op
go test -bench=. -count=10  -run=^$ > new.txt    # 10 runs for benchstat
go test -bench=. -benchtime=2s -run=^$           # run each bench ~2s
go test -bench=. -benchtime=100000x -run=^$      # exactly 100000 iterations
go test -bench=. -cpu=1,4,8 -run=^$              # vary GOMAXPROCS
```

Flag notes:
- `-run=^$` makes the regexp match no normal tests, so only benchmarks run.
- `-benchmem` is cheap insurance — almost always include it.

---

## Reading the output

```
BenchmarkParse-10    932427795    1.230 ns/op    0 B/op    0 allocs/op
            │   │         │           │             │           │
            │   │         │           │             │           └ heap allocations per op
            │   │         │           │             └ bytes allocated per op
            │   │         │           └ time per op (lower better)
            │   │         └ iterations the framework ran (b.N total)
            │   └ GOMAXPROCS the bench ran under
            └ benchmark name
```

- **ns/op** — speed. Lower better.
- **B/op** — memory churn. Lower better.
- **allocs/op** — GC pressure. Lower better; 0 is ideal on a hot path.

Red flag: sub-nanosecond ns/op (e.g. ~0.3) that doesn't grow with more work ⇒
the compiler eliminated your benchmark. Add a sink or use `b.Loop()`.

---

## Hygiene helpers

```go
b.ResetTimer()      // zero timer+counters after setup
b.StopTimer()       // pause timing (e.g. per-iteration setup)
b.StartTimer()      // resume timing
b.ReportAllocs()    // force B/op + allocs/op even without -benchmem
b.SetBytes(n)       // report throughput as MB/s (n bytes processed per op)
var sink T          // package-level: publish results, defeat dead-code elim
```

---

## benchstat workflow (compare fairly)

```bash
go install golang.org/x/perf/cmd/benchstat@latest

go test -bench=. -count=10 -run=^$ > old.txt   # baseline
# ... make your change ...
go test -bench=. -count=10 -run=^$ > new.txt   # candidate
benchstat old.txt new.txt                      # is the delta real?
```

benchstat prints mean ± variation and a **p-value**. p < 0.05 ⇒ the difference
is statistically real; `~` ⇒ indistinguishable from noise. Always compare on the
**same machine, same thermal/power state, back-to-back**.

---

## pprof — CPU and memory

```bash
# CPU profile from a benchmark
go test -bench=BenchmarkX -run=^$ -cpuprofile=cpu.out
go tool pprof cpu.out
#   (pprof)  top            # hottest functions
#   (pprof)  top -cum       # by cumulative time (callers)
#   (pprof)  list FuncName  # line-by-line cost in a function
#   (pprof)  web            # SVG call graph (needs graphviz)
#   (pprof)  png > p.png    # write graph to file

# Memory profile
go test -bench=BenchmarkX -run=^$ -memprofile=mem.out
go tool pprof -alloc_space mem.out    # total bytes allocated
go tool pprof -alloc_objects mem.out  # allocation counts
go tool pprof -inuse_space mem.out    # live (retained) memory

# Interactive web UI (flame graph in the browser)
go tool pprof -http=:0 cpu.out
```

---

## Escape analysis (what goes to the heap)

```bash
go build -gcflags=-m ./...        # print escape + inline decisions
go build -gcflags='-m -m' ./...   # more verbose reasoning
```

Look for `escapes to heap` / `moved to heap` (an allocation) and
`can inline` / `inlining call to` (cheaper calls). Keeping hot-path values on
the stack avoids allocations and GC pressure.

---

## Profiling a live server (one-liner)

```go
import _ "net/http/pprof"   // registers /debug/pprof/ on the default mux
```

```bash
go tool pprof http://host:8080/debug/pprof/profile?seconds=30  # live 30s CPU
go tool pprof http://host:8080/debug/pprof/heap                # heap snapshot
go tool pprof http://host:8080/debug/pprof/goroutine           # goroutine dump
```

Expose this on an internal/admin port only — never the public internet.

---

## Key terms

- **b.N** — iteration count the framework auto-tunes per benchmark; the classic
  loop runs the body `b.N` times.
- **ns/op** — nanoseconds per operation; the headline speed metric (lower
  better).
- **B/op** — bytes allocated on the heap per operation.
- **allocs/op** — number of heap allocations per operation; drives GC pressure.
- **dead-code elimination** — the compiler deleting work whose result is never
  observed; the cause of fake ~0.3 ns/op benchmarks.
- **escape analysis** — compile-time decision of whether a value lives on the
  stack or escapes to the heap; inspect with `-gcflags=-m`.
- **monomorphization** — *n/a in Go*: Go generics use a dictionary/GC-shape
  approach, not per-type code generation (unlike C++/Rust), so you don't get
  Rust-style per-instantiation specialization. Relevant when reasoning about
  generic hot-path costs.
- **pprof** — Go's profiler; collects CPU/heap/goroutine/block profiles,
  inspected with `go tool pprof`.
- **benchstat** — `golang.org/x/perf` tool that compares benchmark runs and
  reports whether a difference is statistically significant.
- **sink var** — a package-level variable you assign benchmark results to so the
  compiler can't eliminate the work.
- **GC pressure** — the load a program puts on the garbage collector; driven by
  allocs/op, and a leading cause of tail-latency spikes under concurrency.

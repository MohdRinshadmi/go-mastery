# Day 10 — Benchmarking, Profiling, and the Phase 2 Capstone

> Mentor note: "Make it work, make it right, make it fast — in that order." Today is the *fast* part, but with a hard rule senior engineers live by: **you do not optimize without measuring.** Your intuition about what's slow is almost always wrong. Go gives you benchmarks and `pprof` so you measure first, then fix the actual hotspot. Today you'll prove a real optimization with numbers, not vibes.

---

## 1. Benchmarks — built into `go test`

A benchmark is a function `BenchmarkXxx(b *testing.B)` in a `_test.go` file. The framework runs your code `b.N` times, auto-tuning `N` until timing is stable.

```go
func BenchmarkSumPrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {       // the loop the framework scales
        s := make([]int, 0, 1000)    // pre-allocated
        for j := 0; j < 1000; j++ {
            s = append(s, j)
        }
    }
}
```

Run:
```bash
go test -bench=. -benchmem
```
`-benchmem` adds allocation stats. Output:
```
BenchmarkSumPrealloc-10    600000   1850 ns/op   8192 B/op   1 allocs/op
BenchmarkSumNaive-10       180000   6100 ns/op  16376 B/op  11 allocs/op
```
Read it: ops, **ns/op** (time per op — lower better), **B/op** (bytes allocated), **allocs/op** (allocation count). The naive version reallocates 11× as the slice grows; pre-allocating is 1 alloc and ~3× faster. *That's* the Day 2 pre-sizing lesson, now measured.

### Benchmark hygiene
- `b.ResetTimer()` after expensive setup so setup isn't counted.
- `b.ReportAllocs()` to always show allocs.
- Assign results to a package-level sink var to stop the compiler optimizing your work away (dead-code elimination):
  ```go
  var sink int
  func BenchmarkX(b *testing.B){ var r int; for i:=0;i<b.N;i++{ r = work() }; sink = r }
  ```

**Senior take:** A benchmark without `-benchmem` hides half the story. Allocations drive GC pressure, and GC is where Go services spend surprise latency under load. Always look at allocs/op.

## 2. Comparing fairly — benchstat
Run a bench multiple times and compare with `benchstat` (from `golang.org/x/perf`): `go test -bench=. -count=10 > old.txt`, change code, `> new.txt`, `benchstat old.txt new.txt`. It tells you if a difference is statistically real or just noise. Don't trust a single run.

## 3. Profiling with pprof

Benchmarks tell you *that* something is slow; profiles tell you *where*.

### CPU profile from a benchmark
```bash
go test -bench=BenchmarkProcess -cpuprofile=cpu.out
go tool pprof cpu.out
# in the pprof prompt:  top   (hottest funcs),  list FuncName  (line-level),  web  (graph)
```

### Memory profile
```bash
go test -bench=. -memprofile=mem.out
go tool pprof -alloc_space mem.out
```

### In a running server (Phase 4+)
Import `net/http/pprof` and it registers `/debug/pprof/` handlers; then `go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30` grabs a live 30s CPU profile from production. This is how you debug a latency spike in a real service.

**Senior take:** The workflow is always: benchmark to confirm there's a problem → `pprof top` to find the hot function → `pprof list` to find the hot line → fix → re-benchmark to confirm the win. Skipping straight to "I think this loop is slow" wastes days.

## 4. What actually makes Go slow (so you know what to look for)
- **Allocations / GC pressure** — the #1 culprit. Reduce with pre-sizing, `sync.Pool` for reusable buffers (Day 29), avoiding `interface{}` boxing in hot loops, and keeping values on the stack (escape analysis: `go build -gcflags=-m` shows what escapes to the heap).
- **Unnecessary copying** of large structs — pass pointers in hot paths.
- **String concatenation in loops** — use `strings.Builder`, not `+=`.
- **Reflection** (`encoding/json` at extreme scale) — codegen if a profile proves it.
- **Lock contention** (Phase 3) — shows up as time in `sync` runtime functions.

## Common mistakes
1. Optimizing without profiling — fixing code that isn't the bottleneck.
2. Micro-benchmarking something the compiler eliminates (no sink var) → fake "0.3 ns/op".
3. Trusting one benchmark run — use `-count` + benchstat.
4. Premature optimization that hurts readability for a path that runs once at startup.
5. Counting setup time in the benchmark (forgot `b.ResetTimer()`).

## Performance mindset
- Optimize the *hot path*, ignore the cold path. 90% of a service's time is in 10% of the code — profile to find which 10%.
- A clear algorithm beats a clever micro-opt: O(n) over O(n²) dwarfs any constant-factor trick.
- Measure in conditions like production (data sizes, concurrency), not toy inputs.

---

## Expert Thinking Mode — "this is slow"

- **Beginner:** "I'll rewrite the part that looks slow."
- **Senior:** "Benchmark to confirm, pprof to locate, fix the real hotspot, re-benchmark to prove the win. Watch allocs/op, not just ns/op."
- **Staff:** "Is this even on the hot path at production scale and concurrency? What's the p99, not the average? Will GC pressure from allocations dominate under load?"
- **Architect:** "Performance is a system property: caching, batching, and the right data store often beat any code-level tuning. I set latency SLOs and use continuous profiling in prod, not one-off local runs."

---

## Real-world use

- **Cloudflare/Uber** run continuous profiling (`pprof`/Pyroscope) in production to catch regressions and find hotspots under real traffic.
- **Benchmarks in CI** guard hot libraries against perf regressions (benchstat gates).
- **`sync.Pool`** for buffer reuse is standard in high-throughput Go (encoding, proxies) to cut GC pressure — Day 29.
- The Go standard library itself ships exhaustive benchmarks; that's the cultural bar.

---

## Interview Questions

1. How do you write and run a benchmark in Go? What does `-benchmem` add?
2. What do ns/op, B/op, and allocs/op mean? Why care about allocs?
3. Why use `b.ResetTimer()` and a sink variable?
4. Walk me through the benchmark→pprof→fix→re-benchmark loop.
5. What's escape analysis and how do you see what escapes to the heap?
6. Name three common causes of slow Go code.
7. Why is "optimize without measuring" an anti-pattern?

---

## Phase 2 Capstone (in `../exercises/` and `../solutions/`)

A **storage abstraction** with two implementations behind one interface — the culmination of Phase 2 (interfaces, composition, DI, generics, testing, benchmarking):

- `Store` interface: `Set(key, value)`, `Get(key) (value, ok)`.
- `MapStore` (a plain `map` + mutex) and `SliceStore` (a linear-scan slice) implementations.
- A **table-driven test** that runs the *same* test suite against *both* implementations (proving interface substitutability).
- **Benchmarks** comparing `Get` on both as N grows — you'll see the map win at scale and *measure* why the slice degrades (O(n) scan). That measurement IS the lesson.

Finish the TODOs, run `go test -bench=. -benchmem ./...`, and bring me the numbers + your read of them. Passing this completes Phase 2; Phase 3 (concurrency) is next.

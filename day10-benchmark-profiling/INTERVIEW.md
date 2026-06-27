# Day 10 Interview Questions — Benchmarking & Profiling

Ten questions with model answers. The first seven are the lesson's; the last
three go a level deeper. Try to answer before expanding each one.

---

### 1. How do you write and run a benchmark in Go? What does `-benchmem` add?

<details>
<summary>Answer</summary>

A benchmark is a function `func BenchmarkXxx(b *testing.B)` in a `_test.go`
file. The framework runs your code `b.N` times, auto-tuning `N` until the timing
is stable enough to be meaningful:

```go
func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {   // or: for b.Loop()  (Go 1.24+)
        parse(input)
    }
}
```

Run with `go test -bench=. -run=^$` (the `-run=^$` skips normal tests). The
`-bench` value is a regexp selecting which benchmarks to run.

`-benchmem` adds **memory** columns: **B/op** (bytes allocated per op) and
**allocs/op** (number of heap allocations per op). Without it you only see
ns/op and miss the allocation story, which is usually what matters under load.
</details>

---

### 2. What do ns/op, B/op, and allocs/op mean? Why care about allocs?

<details>
<summary>Answer</summary>

- **ns/op** — nanoseconds per operation (wall time per `b.N` iteration). Lower
  is better.
- **B/op** — bytes allocated on the heap per operation.
- **allocs/op** — number of distinct heap allocations per operation.

You care about allocations because they drive **GC pressure**. Every heap
allocation is memory the garbage collector must later scan and free; more
allocs/op means more frequent GC, which means CPU spent collecting and, worse,
**tail-latency spikes** (p99) under concurrent load. Two functions with the same
ns/op in a micro-benchmark can behave very differently in production if one
allocates 10x more. Reducing allocs/op (pre-sizing, reuse, stack allocation) is
often the highest-leverage optimization.
</details>

---

### 3. Why use `b.ResetTimer()` and a sink variable?

<details>
<summary>Answer</summary>

**`b.ResetTimer()`** zeroes the elapsed time and allocation counters at the
point you call it. Use it after expensive one-time setup (building a fixture,
opening a file) so that setup isn't counted in your per-op numbers. Without it,
setup cost is smeared across every iteration and inflates ns/op.

**A sink variable** defeats **dead-code elimination**. If your benchmark
computes a value and throws it away, the compiler can prove the work is unused
and delete it — giving a fake ~0.3 ns/op. Assigning the result to a
package-level `var sink T` makes the work observable, so the compiler must keep
it and the timing is honest. (Go 1.24's `b.Loop()` handles this automatically.)
</details>

---

### 4. Walk me through the benchmark→pprof→fix→re-benchmark loop.

<details>
<summary>Answer</summary>

1. **Benchmark** the operation to confirm there really is a performance problem
   and to get a baseline number (with `-benchmem`).
2. **Profile** it: `go test -bench=BenchmarkX -cpuprofile=cpu.out`, then
   `go tool pprof cpu.out`. Use `top` to see the hottest functions and
   `list Func` to drill to the hottest *lines*. (`-memprofile=mem.out` for
   allocations.)
3. **Fix** the actual hotspot the profile pointed to — not what you assumed was
   slow.
4. **Re-benchmark** with `-count` + benchstat to *prove* the win is real and not
   noise.

The discipline is: confirm, locate, fix, prove. Skipping the profile and
guessing is the classic way to waste days optimizing code that wasn't the
bottleneck.
</details>

---

### 5. What's escape analysis and how do you see what escapes to the heap?

<details>
<summary>Answer</summary>

**Escape analysis** is a compile-time analysis that decides whether a value can
live on the **stack** (cheap, freed automatically when the function returns) or
must **escape to the heap** (allocated, and later collected by the GC). A value
escapes when its lifetime can outlive the function — e.g. you return a pointer
to a local, store it in an interface, capture it in a closure that outlives the
call, or its size/shape isn't known at compile time.

You see the decisions with:

```bash
go build -gcflags=-m ./...
```

It prints lines like `moved to heap: x` or `... escapes to heap`. Keeping hot-
path values on the stack (so they *don't* escape) avoids allocations and GC
pressure — a common, measurable optimization.
</details>

---

### 6. Name three common causes of slow Go code.

<details>
<summary>Answer</summary>

1. **Allocations / GC pressure** — the #1 culprit. Caused by not pre-sizing
   slices/maps, boxing values into `interface{}` in hot loops, or values
   escaping to the heap. Fix with pre-sizing, `sync.Pool`, and stack-friendly
   code.
2. **Unnecessary copying / string building** — copying large structs by value in
   hot paths (pass pointers), or `+=` string concatenation in a loop (use
   `strings.Builder`).
3. **Reflection and lock contention** — `encoding/json` reflection at extreme
   scale, or goroutines fighting over a mutex (shows up as time in `sync`
   runtime functions in a profile).

Also acceptable: poor algorithmic complexity (O(n²) where O(n) is possible),
which dwarfs any constant-factor concern.
</details>

---

### 7. Why is "optimize without measuring" an anti-pattern?

<details>
<summary>Answer</summary>

Because your intuition about what's slow is almost always wrong. The time in a
real program is usually in unglamorous places — allocation, a map probe, JSON
decoding, a lock — that you wouldn't guess. If you optimize by eyeballing, you:
(a) waste effort on code that isn't the bottleneck, (b) often hurt readability
and introduce bugs, and (c) can't prove your change helped. Measuring first
(benchmark + pprof) tells you *where* the time actually goes; measuring after
(benchstat) tells you whether your fix was a real win or just noise. "Make it
work, make it right, make it fast — and you do not make it fast without
measuring."
</details>

---

### 8. What is `b.Loop()` and why is it better than the `b.N` loop? (Go 1.24+)

<details>
<summary>Answer</summary>

`b.Loop()` is a benchmark loop form added in Go 1.24:

```go
func BenchmarkX(b *testing.B) {
    for b.Loop() {
        sink = work()
    }
}
```

It is better than the classic `for i := 0; i < b.N; i++` for two reasons:

1. **It prevents dead-code elimination of the loop body** and constant-folding
   of loop-invariant inputs. The compiler is told not to optimize the body away,
   so you get honest numbers without the manual package-level sink trick.
2. **Setup/teardown outside the loop runs exactly once** and is automatically
   excluded from timing, so you don't need `b.ResetTimer()` for the common case.

It's now the recommended form for new benchmarks. (The classic `b.N` loop is
still valid and is what you must use if you want to *demonstrate* dead-code
elimination — which is exactly why this Day's debugging challenge uses it.)
</details>

---

### 9. Explain stack vs heap and what escape analysis tells you about a value.

<details>
<summary>Answer</summary>

The **stack** is per-goroutine memory that grows and shrinks with function
calls. Allocating on the stack is nearly free (just bump a pointer) and the
memory is reclaimed automatically when the function returns — the GC never sees
it. The **heap** is shared memory for values whose lifetime outlives the
function that created them; heap allocations cost more and must later be found
and freed by the **garbage collector**.

**Escape analysis** tells you which side a given value lands on. If the compiler
can prove a value's lifetime is bounded by its function, it stays on the stack.
If the value might be referenced after the function returns — returned by
pointer, stored in an interface, captured by an escaping closure, or of a size
not known at compile time — it "escapes" to the heap. The practical payoff:
fewer escapes means fewer allocations means less GC pressure. You inspect the
decisions with `go build -gcflags=-m`.
</details>

---

### 10. How do you profile a live production server, and what does benchstat's p-value mean?

<details>
<summary>Answer</summary>

**Profiling a live server:** import the side-effect package
`net/http/pprof` (`import _ "net/http/pprof"`). It registers handlers under
`/debug/pprof/` on the default mux. Then, from your laptop, pull a profile from
the running process:

```bash
go tool pprof http://host:8080/debug/pprof/profile?seconds=30   # 30s CPU profile
go tool pprof http://host:8080/debug/pprof/heap                 # heap snapshot
```

This captures real production behavior — real data sizes, real concurrency —
which a local micro-benchmark can't reproduce. (Expose the endpoint only on an
internal/admin port; never to the public internet.) This is how you diagnose a
latency spike under live traffic.

**benchstat's p-value:** when comparing two sets of runs, benchstat performs a
statistical test and reports a **p-value** — the probability that the observed
difference between old and new could have arisen from random noise alone if
there were truly no difference. A small p-value (commonly < 0.05) means the
difference is **statistically significant** — likely a real change, not noise. A
large p-value (or benchstat printing `~`) means the runs are indistinguishable;
your "improvement" can't be trusted. This is why you never conclude from a
single run.
</details>

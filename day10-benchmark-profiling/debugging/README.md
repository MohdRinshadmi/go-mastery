# Day 10 Debugging Challenge — The Benchmark the Compiler Optimized Away

A benchmark is only useful if it actually runs the code you think it runs. This
one doesn't. The fastest way to ship a "10x faster!" lie is to write a benchmark
the compiler quietly deletes.

## Symptom

`bugged/` reports **~0.3 ns/op** — too good to be true. At ~0.3 ns you are well
under the cost of a single multiply on a modern CPU, so the benchmark cannot be
doing the arithmetic it claims. Worse: the number **does not scale at all** with
how much work `sumSquares` does. Make the function ten times heavier and the
reported ns/op barely moves. That is the tell.

## Repro

Run both modules and compare:

```bash
cd /Users/ioss/Documents/StudyProjects/GO/day10-benchmark-profiling/debugging/bugged
go test -bench=. -benchmem -run=^$ ./...

cd /Users/ioss/Documents/StudyProjects/GO/day10-benchmark-profiling/debugging/fixed
go test -bench=. -benchmem -run=^$ ./...
```

What you'll see (Apple M4 — your numbers will differ, the *ratio* is the point):

```
# bugged  -> the lie
BenchmarkSumSquares-10    1000000000    0.2380 ns/op    0 B/op    0 allocs/op

# fixed    -> the truth
BenchmarkSumSquares-10     932427795    1.230  ns/op    0 B/op    0 allocs/op
```

The only difference between the two modules is **one line in `sum_test.go`**:
whether the result of `sumSquares` is used or thrown away. `sum.go` is identical
in both. The function was never the problem — the benchmark was.

## Hint

- The discarded result is the clue. What can the compiler prove about a *pure*
  function whose return value nobody reads?
- Look at what `bugged/sum_test.go` does with `sumSquares(arg)` versus what
  `fixed/sum_test.go` does (`r = sumSquares(arg)` ... `sink = r`).
- Try `go build -gcflags=-m ./...` and notice `sumSquares` is inlinable. Once
  inlined into the loop, its output is dead — so the optimizer removes it.

<details>
<summary>Solution &amp; why</summary>

### What's happening: dead-code elimination

`sumSquares` is **pure** (no side effects) and small enough that the compiler
**inlines** it into the benchmark loop. After inlining, the bugged benchmark is:

```go
for i := 0; i < b.N; i++ {
    // ...a block of arithmetic whose result is never read...
}
```

The optimizer can prove that nothing observes the computed value, so the whole
block is **dead code** and gets deleted. You are then timing an empty loop —
hence the sub-nanosecond ~0.3 ns/op, and hence the fact that it never changes
when you add more work to `sumSquares`. The benchmark is measuring nothing.

### The fix: a package-level sink

Make the result **observable** outside the loop by assigning it to a
package-level variable:

```go
var sink int

func BenchmarkSumSquares(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = sumSquares(arg)
    }
    sink = r // publish: now the work is observable and cannot be deleted
}
```

Because `sink` is a package-level variable, the compiler must assume something
else might read it, so it can no longer prove the work is dead. The call runs
every iteration and the timing becomes honest (~1.2 ns/op here). A local `_ = `
is **not** reliably enough — the compiler can still see a local is never read.
The result must escape the function's view, which a package-level var does.

> We also keep the argument in a package-level `var arg = 1000` rather than a
> literal. Otherwise `sumSquares(1000)` is a compile-time constant the optimizer
> could fold to a single number, which would hide the real per-call cost in the
> fixed version too.

### Modern fix: `for b.Loop()` (Go 1.24+)

Go 1.24 added a benchmark loop form that solves this for you:

```go
func BenchmarkSumSquares(b *testing.B) {
    for b.Loop() {
        sink = sumSquares(arg)
    }
}
```

`b.Loop()` is implemented so the compiler is forbidden from optimizing the loop
body away (and it also stops constant-folding the inputs), so you get honest
numbers without the manual sink dance. It is the recommended form for new
benchmarks. We used the classic `b.N` loop in this challenge *specifically*
because `b.Loop()` would have prevented the bug — defeating the lesson.
</details>

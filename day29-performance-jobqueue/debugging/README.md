# Day 29 debugging — the hot path that allocates a buffer every call

**Phase 6 · performance · allocations & `sync.Pool`**

> Stdlib only. The "fix" is a behavior/benchmark difference you can measure with
> `-benchmem`, not a crash.

## Symptom

A `Render` function formats a record into bytes. It's on a very hot path — called
for every job/event/log line. Under load, the allocation profile (`pprof -alloc_objects`)
shows `Render` near the top and GC time is a measurable slice of CPU. The function
is correct; it's just allocating far more than it needs to.

```bash
cd bugged
go test -bench=. -benchmem    # look at the allocs/op column
go run .                       # quick allocs-per-record estimate
```

Expected (after the fix): roughly **1 alloc/op** (only the result the caller keeps).
Actual (bugged): **2 allocs/op** — a fresh `bytes.Buffer` *plus* the result, every call.

## Hint

What lives only for the duration of one call but gets created on every single call?
A short-lived, frequently-allocated object on a hot path is the textbook case for
one specific `sync` primitive. Profile first (that's the rule), then ask: can this
object be *reused* instead of *reallocated*?

## How to reproduce

`go test -bench=. -benchmem` in both `bugged/` and `fixed/` and compare the
`allocs/op` and `B/op` columns. Or `go run .` in each for an allocs-per-record count.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

`Render` allocates a brand-new `bytes.Buffer` on every call:

```go
func Render(r Record) []byte {
    var buf bytes.Buffer // heap allocation, every single call
    // ... write into buf ...
    return append([]byte(nil), buf.Bytes()...)
}
```

The buffer escapes to the heap (it's used to build a return value) and is thrown
away immediately — so each call creates garbage the GC must later collect. Called
millions of times per second, that allocation churn becomes a top entry in the
alloc profile and shows up as GC CPU cost. Reducing allocations is the single
biggest lever for most Go services, and this is the canonical shape of the problem.

### The fix

The buffer is short-lived and allocated constantly — exactly what **`sync.Pool`**
is for. Reuse buffers instead of reallocating them:

```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func Render(r Record) []byte {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()            // CRUCIAL — clear leftover bytes from the last use
    defer bufPool.Put(buf)
    // ... write into buf ...
    return append([]byte(nil), buf.Bytes()...) // copy out; buffer returns to pool
}
```

The per-call buffer allocation disappears; only the result copy (which the caller
keeps) remains. On the benchmark that's 2 allocs/op → 1 alloc/op, and the saved
allocations are pure GC relief on the hot path.

Two things that make `sync.Pool` correct rather than a footgun:

> 1. **Always `Reset()`** the object after `Get` (or before reuse). Forgetting it
>    leaks the previous user's data into the next call — a real correctness bug.
> 2. **Copy out anything the caller keeps.** Returning `buf.Bytes()` directly would
>    hand out memory that's about to be reused by another goroutine — a data race.

And the discipline around it:

> - `sync.Pool` is a **precision tool, not a default**. Reach for it only when a
>   profile shows a hot object allocated millions of times. Pooling long-lived or
>   rarely-allocated objects adds bugs for no gain.
> - It's GC-aware (entries can be cleared on GC), so it's a *cache*, not a guarantee.
> - Other levers for the same goal: pre-size slices/maps (`make([]T, 0, n)`),
>   `strings.Builder` over `+=`, avoid `interface{}` boxing in hot loops.

`fixed/` reuses buffers (fewer allocs/op); the bugged version allocates one per call.

</details>

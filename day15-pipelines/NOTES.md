# Day 15 — Pipelines & Cancellation: Quick Reference

## Pipeline stage shape

```go
func stage(ctx context.Context, in <-chan T) <-chan U {
    out := make(chan U)
    go func() {
        defer close(out)              // close cascades downstream
        for v := range in {           // ends when in is closed & drained
            select {
            case out <- transform(v): // normal send
            case <-ctx.Done():        // early teardown -> exit, no leak
                return
            }
        }
    }()
    return out
}
```

Source stage:
```go
func gen(ctx context.Context, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

## Two disciplines (need both)

| Discipline | Purpose | Mechanism |
|---|---|---|
| **Close** (forward) | normal "no more data" | `defer close(out)` per stage → `range` ends, cascades |
| **Cancellation** (everywhere) | early/abnormal "stop now" | `select` on `ctx.Done()` in every send; consumer `cancel()`s |

## Consumer that stops early

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()                        // safety net for every return path
for v := range stage(ctx, gen(ctx, ...)) {
    use(v)
    if enough { cancel(); break }     // tear down upstream promptly
}
```

## Parallelize a slow stage
Fan-out the stage into M workers over the same `in`, fan-in their outputs (Day 14
merge): `WaitGroup` + `go func(){ wg.Wait(); close(out) }()`.

## Leak test
```go
base := runtime.NumGoroutine()
// run pipeline with an early break
time.Sleep(...) // let cancelled stages unwind
if runtime.NumGoroutine() > base { /* LEAK */ }
// or use go.uber.org/goleak in tests
```

## Bounded memory
Unbuffered channels → ≤1 item per stage in flight → stream unbounded input in
fixed memory. Buffer deliberately (profile) to smooth bursty stages.

---

## Key terms

- **Pipeline** — stages connected by channels, each a goroutine reading→writing.
- **Stage** — `func(ctx, in <-chan T) <-chan U`; owns one goroutine.
- **Close cascade** — `defer close(out)` ending each downstream `range` in turn.
- **Cancellation propagation** — `select` on `ctx.Done()` so early exit tears down all stages.
- **Backpressure** — slow downstream stalls upstream (natural with unbuffered channels).
- **Fan-out / fan-in** — parallelize a slow stage and re-merge its outputs.
- **Bounded memory** — ≤1 item per stage in flight (unbuffered).
- **goleak** — test helper that fails if goroutines outlive the test.

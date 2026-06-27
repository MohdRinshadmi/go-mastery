# Day 15 debugging — the pipeline that leaks when you leave early

**Phase 3 · Concurrency · pipeline stage with no cancellation path**

## Symptom

A streaming pipeline `gen → square`. The consumer only wants the first 3 results,
so it `break`s out of the range early — a totally reasonable thing to do. The
results are correct, but every early exit **leaks the upstream goroutines**. Run
it:

```bash
cd bugged
go run .
```

```
got: 1
got: 4
got: 9
after taking 3 results: 2 goroutines leaked (gen + square stranded)
GOROUTINE LEAK: gen blocked on `out <- i`, square blocked on `out <- n*n`
  the consumer's early break never told upstream to stop
```

In a server, every request that takes "just the first few" results would strand
two goroutines forever. That's a slow memory death.

## Hint

The Day 15 leak test: **"if my consumer returns early, does every upstream
goroutine exit?"**

When the consumer `break`s, who is still trying to send? `square` is blocked on
`out <- n*n` (no receiver anymore), and `gen` is blocked on its send into
`square`. Nothing tells them to stop. What primitive lets a downstream signal
"I'm done, tear yourself down" to every upstream stage?

## How to reproduce

`go run .` in `bugged/` — `runtime.NumGoroutine()` stays elevated (2 above
baseline) after the consumer finishes, instead of returning to baseline.
Deterministic every run.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

No stage has a **cancellation path**. Each stage only knows how to send:

```go
for n := range in {
    out <- n * n   // blocks forever once the consumer stops receiving
}
```

When the consumer `break`s, it stops receiving from `square`. `square`'s next
`out <- n*n` blocks forever (no receiver). That means `square` stops receiving
from `gen`, so `gen`'s `out <- i` blocks forever too. Both goroutines are
stranded — a goroutine leak. The consumer simply returning does **not** propagate
upward; channels don't have a "the reader left" signal on their own.

### The fix

Thread a `context.Context` through every stage and `select` on `ctx.Done()` for
every send. The consumer cancels when it leaves:

```go
func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:    // normal send
            case <-ctx.Done():    // downstream gave up -> exit, no leak
                return
            }
        }
    }()
    return out
}

// consumer:
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
for v := range square(ctx, gen(ctx)) {
    ...
    if enough { cancel(); break }   // tears down the whole pipeline
}
```

`cancel()` closes `ctx.Done()`, which every stage is selecting on. Each blocked
send loses the race to `ctx.Done()`, the goroutine returns, `defer close(out)`
runs, and the closes cascade. `fixed/` returns to baseline — 0 leaked — and is
clean under `-race`.

### The rules

> 1. **Every send and receive in a long-lived pipeline `select`s on
>    `ctx.Done()`.** A send with no cancellation path is a latent leak the moment
>    a consumer exits early.
> 2. **The consumer owns the `cancel`** — `defer cancel()` guarantees teardown on
>    every return path (early break, error, or normal completion).
> 3. **Test the leak directly:** assert `runtime.NumGoroutine()` returns to
>    baseline, or use `go.uber.org/goleak`.

Verify: `go run -race .` in `fixed/` reports 0 leaked goroutines.

</details>

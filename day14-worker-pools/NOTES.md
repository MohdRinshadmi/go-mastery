# Day 14 — Worker Pools, Fan-Out/In, errgroup: Quick Reference

## Bounded worker pool

```go
jobs := make(chan int)
results := make(chan int)

// producer
go func() {
    for _, j := range work { jobs <- j }
    close(jobs)                 // stops workers' range jobs
}()

// pool of N workers (fan-out)
var wg sync.WaitGroup
for w := 0; w < N; w++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs { results <- process(j) }
    }()
}

// one coordinator closes results after ALL workers finish
go func() { wg.Wait(); close(results) }()

// consumer
for r := range results { collect(r) }
```

**Close discipline**
- Producer closes `jobs` (stops workers).
- Coordinator closes `results` after `wg.Wait()` (stops consumer).
- Sender closes, once. Receiver never closes. Multiple senders → coordinator.

**Pool size N**
- CPU-bound → ~`runtime.NumCPU()`.
- I/O-bound → higher, capped by **downstream capacity**. Benchmark.

## Fan-in (merge channels)

```go
func fanIn(cs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, c := range cs {
        wg.Add(1)
        go func(c <-chan int) { defer wg.Done(); for v := range c { out <- v } }(c)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}
```

## errgroup (the modern default)

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx) // first error cancels ctx
g.SetLimit(n)                        // bound concurrency = pool with errors for free
for _, item := range items {
    item := item
    g.Go(func() error { return work(ctx, item) }) // select on ctx.Done() inside
}
err := g.Wait()                      // first non-nil error, after all return
```

- First task error → cancels the shared `ctx` → siblings bail via `ctx.Done()`.
- `g.Wait()` returns the **first** error.
- `SetLimit(n)` = bounded pool + error handling + cancellation.

## Leak prevention
- Every `results <- x` in a long-lived worker should `select` on `ctx.Done()`.
- Close `jobs` (else workers + `wg.Wait()` hang) and `results` (else consumer hangs).

---

## Key terms

- **Worker pool** — fixed N goroutines pulling from a shared `jobs` channel.
- **Bounded concurrency** — capping in-flight work to protect memory & downstream.
- **Fan-out** — many goroutines reading one channel (distributing work).
- **Fan-in** — merging many channels into one (collecting results).
- **Coordinator close** — `go func(){ wg.Wait(); close(out) }()`.
- **errgroup** — `x/sync` group: bounded, error-capturing, context-cancelling.
- **SetLimit** — errgroup's concurrency bound.
- **Backpressure** — slowing the producer when consumers can't keep up.
- **Downstream capacity** — the real constraint on pool size for I/O work.

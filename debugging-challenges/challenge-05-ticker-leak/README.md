# Challenge 05 — the service that leaks a goroutine per request

**Phase 5 · Production engineering · resource leaks**

## Symptom

A request handler spins up a small background poller (a `time.Ticker` in a goroutine) to refresh some per-request state, and tears it down when the request finishes. Memory and the goroutine count look fine in a quick local test.

In production, the goroutine count climbs forever and never comes back down. After a few hours the process is holding tens of thousands of idle goroutines and a matching pile of unstopped tickers. Classic slow leak. This program simulates many requests and prints the goroutine count before and after:

```bash
cd bugged
go run .
```

Expected: goroutine count returns to roughly its starting value after all requests finish.
Actual: it stays elevated — one leaked goroutine (and one un-stopped ticker) per request.

## Hint

Two separate leaks hide here, and they're cousins:

1. `time.NewTicker` allocates a runtime timer that keeps firing until you call `.Stop()`. Did every path call it?
2. A goroutine that loops on `for range ticker.C { ... }` with no exit condition **never returns**. Even if you stop the ticker, a goroutine blocked forever on a channel that no longer receives is a leak. What signal tells the goroutine to return?

`defer` and a `done`/`ctx.Done()` channel are the tools. Run with `go run .` and watch `runtime.NumGoroutine()`.

## How to reproduce

`go run .` in `bugged/`. It launches 100 short "requests", waits, forces a GC, and reports the goroutine delta.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The buggy `startPoller` has two defects:

```go
func startPoller() {
    ticker := time.NewTicker(10 * time.Millisecond)
    go func() {
        for range ticker.C { // never returns
            _ = doRefresh()
        }
    }()
    // no ticker.Stop(), no way to stop the goroutine
}
```

1. **Ticker never stopped.** `time.NewTicker` registers a timer with the runtime that fires forever. Without `ticker.Stop()` the timer (and its memory) lives until process exit. The docs are explicit: *"The Ticker must be stopped to release associated resources."*
2. **Goroutine never exits.** `for range ticker.C` only ends when `ticker.C` is closed — but `Stop()` does **not** close the channel. So even after a `Stop()`, the goroutine is parked forever waiting on a channel that will never deliver again. That parked goroutine, plus everything it closes over, is pinned for the life of the process.

Per request, you leak one goroutine and one ticker. Multiply by your request rate × uptime and you get the production graph that only goes up.

### The fix

Give the goroutine an explicit exit signal and stop the ticker on the way out:

```go
func startPoller() (stop func()) {
    ticker := time.NewTicker(10 * time.Millisecond)
    done := make(chan struct{})

    go func() {
        defer ticker.Stop() // release the runtime timer
        for {
            select {
            case <-ticker.C:
                _ = doRefresh()
            case <-done: // explicit exit -> goroutine returns
                return
            }
        }
    }()

    return func() { close(done) }
}
```

Caller:

```go
stop := startPoller()
defer stop() // guaranteed teardown on every path
```

The `select` lets the goroutine choose between "do work" and "go home." Closing `done` makes `<-done` ready, the goroutine `return`s, and the deferred `ticker.Stop()` frees the timer. Net leak per request: zero. (A `context.Context` works identically — `case <-ctx.Done(): return` — and is the idiom when the lifetime is already tied to a request context, as in Challenge 04.)

Rules:

> 1. Every `time.NewTicker` / `time.NewTimer` you don't let run for the whole program needs a `Stop()`, ideally `defer`red right after creation.
> 2. Every goroutine must have a *reachable* exit. "Loops forever on a channel" is not an exit. Give it a `done`/`ctx.Done()` case.
> 3. Test for leaks: snapshot `runtime.NumGoroutine()` before and after; it should return to baseline.

`fixed/` shows the goroutine count returning to baseline after all requests complete.

</details>

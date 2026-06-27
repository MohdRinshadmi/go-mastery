# Challenge 04 — the request that won't quit

**Phase 4 · Backend · context & cancellation**

## Symptom

An HTTP handler calls a slow downstream "database" lookup. When the client disconnects (closes the tab, times out, hits Ctrl-C on curl), the work should be abandoned — the handler should notice cancellation and stop.

It doesn't. The downstream call runs to completion every time, burning the full delay even though nobody is waiting for the answer. Under load this pins goroutines and CPU on work that will be thrown away. This program simulates it with a self-cancelling client:

```bash
cd bugged
go run .
```

Expected output: the handler returns *quickly* with a cancellation message (~50ms).
Actual: it blocks the full 2s and logs that it finished work for a client that's already gone.

## Hint

`context.Background()` is a *root* context — it is never cancelled. Where should the handler's context come from instead? Look at what `r.Context()` gives you, and what `context.WithTimeout` / `WithCancel` derive *from*. If your downstream call's context isn't a descendant of the request's context, the request's cancellation can't reach it.

## How to reproduce

`go run .` in `bugged/`. The harness fires a request, cancels the client after 50ms, and times how long the handler keeps working.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The handler builds its timeout context from `context.Background()`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

`context.Background()` has **no parent that can be cancelled**. So when the client disconnects, `net/http` cancels `r.Context()` — but that signal goes nowhere, because the handler's `ctx` is not derived from `r.Context()`. The downstream `queryDB(ctx, ...)` keeps running until its own 5s timeout (or completion), oblivious to the fact that the caller left.

Two failures often travel together here:
1. Not deriving from `r.Context()` — so client disconnects don't propagate.
2. Forgetting to `defer cancel()` on the derived context — leaking the context's timer/goroutine until it fires.

### The fix

Derive the handler's context from the **request** context, and always `defer cancel()`:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel() // release the timer; also cancels children on return

    result, err := queryDB(ctx, "SELECT ...")
    if err != nil {
        // ctx.Err() is context.Canceled when the client left,
        // context.DeadlineExceeded when our 5s budget blew.
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    fmt.Fprintln(w, result)
}
```

Now `r.Context()` cancellation flows down: client disconnect → `r.Context()` cancelled → derived `ctx` cancelled → `queryDB` sees `<-ctx.Done()` and bails immediately. And `queryDB` itself must actually *select on* `ctx.Done()` rather than blindly `time.Sleep`-ing — a downstream that ignores its context can't be cancelled no matter how clean the plumbing above it is.

Rules:

> 1. In an HTTP handler, the request's context is the root of all the work that request spawns. Derive from `r.Context()`, never from `context.Background()`.
> 2. Every `WithTimeout`/`WithCancel`/`WithDeadline` gets a `defer cancel()` — no exceptions. `go vet` will warn when you forget.
> 3. A context is only as useful as the code that checks it. Blocking calls must `select` on `ctx.Done()`.

`fixed/` shows the timed harness returning in ~50ms with `context canceled`, instead of grinding through the full delay.

</details>

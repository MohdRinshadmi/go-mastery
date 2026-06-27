# Day 25 Notes — Rate Limiting & Graceful Shutdown (quick reference)

## Token bucket
- Bucket holds up to `burst` tokens, refills at `rate` tokens/sec.
- Each request takes a token; none → reject (or wait).
- `rate.NewLimiter(rate.Limit(10), 20)` = 10 rps sustained, bursts to 20.

```go
if !limiter.Allow() {                 // non-blocking reject
    w.Header().Set("Retry-After", "1")
    http.Error(w, "rate limited", http.StatusTooManyRequests) // 429
    return
}
// limiter.Wait(ctx) instead = BLOCK until a token (smoothing)
```

## Allow vs Wait
| Method | Behavior | Use |
|---|---|---|
| `Allow()` | non-blocking, false if empty | reject abusive inbound (429) |
| `Wait(ctx)` | blocks for a token | smooth internal/outbound calls |

## Per-client limiting
- Keep a `map[key]*rate.Limiter` (per IP / API key / user) under a mutex.
- **Evict idle entries** or it leaks memory.
- In-memory = per-pod. For a true global cap use Redis (shared counter).

## Response conventions
- **429 Too Many Requests** + `Retry-After` + `X-RateLimit-Remaining`.
- Pair server limits with client backoff + jitter.

## Graceful shutdown (Go 1.16+)
```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
srv := &http.Server{Addr: ":8080", Handler: mux}
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()
<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
_ = srv.Shutdown(shutdownCtx)   // NOT srv.Close()
```

## Shutdown(ctx) does
1. Stop accepting new connections.
2. Wait for in-flight requests up to the deadline.
3. Return when drained (or forced by the deadline).

Rules: **wait for it to return** · deadline **< grace period** · readiness **off
first** · then drain workers / close DB / flush traces.

## Production hardening checklist
- `http.Server` timeouts: `ReadTimeout`, `WriteTimeout`, `IdleTimeout`,
  `ReadHeaderTimeout` (the zero value has NONE).
- `http.MaxBytesReader` for body-size DoS.
- Panic-recovery middleware (per request).
- `MaxHeaderBytes`, security headers, TLS, no secrets in logs.

## Key terms
- **Token bucket** — burst-tolerant rate algorithm (rate + burst).
- **429 / Retry-After** — rate-limit response + back-off hint.
- **`signal.NotifyContext`** — ctx cancelled on SIGINT/SIGTERM.
- **`Shutdown` vs `Close`** — drain vs immediate sever.
- **Grace period** — orchestrator SIGTERM→SIGKILL window (K8s default 30s).
- **Slow-loris** — DoS via trickled bytes on a timeout-less server.

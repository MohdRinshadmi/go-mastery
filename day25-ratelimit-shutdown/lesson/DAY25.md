# Day 25 — Rate Limiting, Graceful Shutdown & Production Hardening

> Mentor note: This is the day your service learns to survive contact with the real world. **Rate limiting** stops one abusive client (or a retry storm) from taking everyone down. **Graceful shutdown** means a deploy doesn't drop in-flight requests on the floor — the difference between "zero-downtime deploy" and "every deploy causes a blip of 502s." Both are small amounts of code with outsized impact on reliability. This closes Phase 5: your service is now observable (metrics/traces/health) *and* operable.

---

## 1. Rate limiting

### Why
- Protect the service from overload (a buggy client looping, a scraper, a DDoS).
- Protect downstreams (your DB) from being hammered.
- Enforce fairness / business tiers (free vs paid quotas).

### The token bucket (the standard algorithm)
A bucket holds up to `burst` tokens, refilled at `rate` tokens/sec. Each request takes a token; no token → rejected (or it waits). This allows short bursts but bounds the sustained rate. Go's `golang.org/x/time/rate` implements it:

```go
import "golang.org/x/time/rate"

// 10 requests/sec sustained, bursts up to 20
limiter := rate.NewLimiter(rate.Limit(10), 20)

if !limiter.Allow() {                 // non-blocking: reject if no token
    http.Error(w, "rate limited", http.StatusTooManyRequests) // 429
    return
}
// limiter.Wait(ctx) instead would BLOCK until a token (with ctx timeout)
```

### Per-client limiting
A global limiter protects the service but lets one client starve others. For fairness, keep a **per-key limiter** (per IP / API key / user):

```go
type ipLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
}
func (l *ipLimiter) get(ip string) *rate.Limiter {
    l.mu.Lock()
    defer l.mu.Unlock()
    if lim, ok := l.limiters[ip]; ok { return lim }
    lim := rate.NewLimiter(10, 20)
    l.limiters[ip] = lim
    return lim
}
```
(Real systems evict idle entries to bound memory, and use Redis for limits shared across instances — a single-instance map only limits per-pod.)

### Response conventions
- Return **429 Too Many Requests**.
- Add `Retry-After` and `X-RateLimit-Remaining` headers so well-behaved clients back off.

**Senior take:** A global in-memory limiter is a starting point, not the answer. Across N replicas behind a load balancer, each has its own bucket — your real limit is N× what you set. For a true global limit, the counter must live in a shared store (Redis). Know which one you've actually built. And always pair server-side limits with client-side backoff+jitter (Day 1's retry lesson) — they're two halves of the same problem.

## 2. Graceful shutdown

### The problem
On deploy, the orchestrator sends `SIGTERM`, then `SIGKILL` after a grace period. If you exit immediately on SIGTERM, every in-flight request is severed → 502s. Graceful shutdown: **stop accepting new connections, let in-flight requests finish (up to a deadline), then exit.**

### The pattern (Go 1.16+)
```go
func main() {
    // ctx is cancelled when SIGINT/SIGTERM arrives
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":8080", Handler: mux}

    // run the server in a goroutine so main can wait for the signal
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    <-ctx.Done() // block until a shutdown signal
    log.Println("shutting down...")

    // give in-flight requests up to 15s to finish, then force-close
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Println("forced shutdown:", err)
    }
    log.Println("stopped cleanly")
}
```

`srv.Shutdown(ctx)`:
1. Stops listening (no new connections).
2. Waits for active requests to complete, up to the ctx deadline.
3. Returns when done (or the deadline forces it).

### Don't forget the rest of the cleanup
Also drain on shutdown: flush logs/traces, close the DB pool, finish in-flight background jobs/workers (cancel their context, `wg.Wait()`), deregister from service discovery. Order matters: stop intake → drain → close resources.

**Senior take:** The shutdown deadline must be **shorter than the orchestrator's grace period** (K8s `terminationGracePeriodSeconds`, default 30s) or you get SIGKILLed mid-drain — worst of both worlds. And readiness (Day 24) should flip to "not ready" *before* you start draining, so the load balancer stops sending new traffic a beat before you stop accepting it. Graceful shutdown + readiness gating = true zero-downtime deploys.

## 3. Other production hardening (the checklist)
- **Server timeouts** — `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout` on `http.Server`. Without these, a slow-loris client can tie up connections forever. **The default `http.Server{}` has NO timeouts** — always set them.
- **Body size limits** — `http.MaxBytesReader` to stop giant-payload DoS.
- **Panic recovery middleware** — recover per request (Day 4) so one panic doesn't kill the server.
- **Security headers**, TLS, no secrets in logs, dependency scanning.
- **`MaxHeaderBytes`**, connection limits.

## Common mistakes
1. `http.Server{}` with no timeouts → slow-loris / resource exhaustion.
2. Exiting on SIGTERM without `Shutdown` → dropped requests on every deploy.
3. Shutdown deadline ≥ orchestrator grace period → SIGKILL mid-drain.
4. Global in-memory rate limiter assumed to be cluster-wide (it's per-pod).
5. Rate limiting without 429 + Retry-After → clients can't back off correctly.
6. Forgetting to drain background workers / close DB on shutdown.
7. Per-IP limiter map that never evicts → memory leak.

## Performance
- `rate.Limiter` is cheap and lock-free-ish; per-key maps need a mutex — shard or use `sync.Map` at very high QPS.
- Graceful shutdown adds no steady-state cost; it only affects the shutdown window.

---

## Expert Thinking Mode — "make it production-ready"

- **Beginner:** "It serves requests, ship it."
- **Senior:** "Timeouts on the server, panic recovery, rate limit with 429+Retry-After, graceful shutdown that drains in-flight, readiness flips before drain."
- **Staff:** "Cluster-wide limits via Redis; backpressure and load shedding under overload; shutdown ordering (readiness → drain HTTP → drain workers → close DB) within the grace period; chaos-test the deploy."
- **Architect:** "Reliability is systemic: limits + backoff + circuit breakers + bulkheads across services; zero-downtime rollout strategy; capacity and overload behavior are designed, not hoped for."

---

## Real-world use

- **Every public API** rate-limits (Stripe, GitHub return 429 + headers). Cloudflare's edge is essentially planet-scale rate limiting.
- **Graceful shutdown** is mandatory for zero-downtime deploys on Kubernetes; the readiness-then-drain dance is standard.
- **Missing server timeouts** is a recurring real CVE/incident cause in Go services.

---

## Interview Questions

1. Explain the token bucket algorithm. What do `rate` and `burst` control?
2. Why is a global in-memory rate limiter wrong across multiple replicas? Fix?
3. What status code and headers should a rate-limited response use?
4. Walk through graceful shutdown with `signal.NotifyContext` + `srv.Shutdown`.
5. Why must the shutdown deadline be shorter than the orchestrator grace period?
6. Why flip readiness to "not ready" before draining?
7. What happens if you use `http.Server{}` with no timeouts?

---

## Your tasks

`../exercises/`: (1) add a per-IP rate-limiting middleware returning 429, (2) configure an `http.Server` with sane timeouts, and (3) implement graceful shutdown with `signal.NotifyContext` + `srv.Shutdown` that drains in-flight requests. The runnable demo lets you `curl` fast to trigger 429 and Ctrl-C to watch a clean drain. Reference in `../solutions/`. Passing this completes Phase 5 — your service is production-grade.

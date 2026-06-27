# Day 25 Interview Questions — Rate Limiting & Graceful Shutdown

Lesson questions plus extras. Answers in `<details>`.

---

### 1. Explain the token bucket. What do `rate` and `burst` control?

<details>
<summary>Answer</summary>

A bucket holds up to `burst` tokens and refills at `rate` tokens/sec. Each request
takes a token; if none is available it's rejected (or waits). `burst` is the
maximum short spike allowed (the bucket capacity); `rate` is the sustained
throughput over time. This permits brief bursts while bounding the long-run rate.
`rate.NewLimiter(rate.Limit(10), 20)` = 10 rps sustained, bursts up to 20.
</details>

---

### 2. Why is a global in-memory rate limiter wrong across replicas? Fix?

<details>
<summary>Answer</summary>

Each replica has its own in-memory bucket, so with N replicas behind a load
balancer the effective limit is N× your configured value. For a true global limit
the counter must live in a shared store (Redis) so all instances share one bucket.
In-memory is fine for per-pod fairness, not a cluster-wide cap.
</details>

---

### 3. What status code and headers should a rate-limited response use?

<details>
<summary>Answer</summary>

**429 Too Many Requests**, with `Retry-After` (when to retry) and
`X-RateLimit-Remaining`/`X-RateLimit-Limit` so well-behaved clients can back off
correctly. Pair with client-side backoff + jitter.
</details>

---

### 4. Walk through graceful shutdown with `signal.NotifyContext` + `srv.Shutdown`.

<details>
<summary>Answer</summary>

`ctx, stop := signal.NotifyContext(ctx, SIGINT, SIGTERM)` makes `ctx` cancel on a
signal. Run `srv.ListenAndServe()` in a goroutine; block on `<-ctx.Done()`. On
signal, create a deadline context and call `srv.Shutdown(shutdownCtx)` — it stops
accepting new connections, drains in-flight requests up to the deadline, then
returns. Wait for it to return before exiting, and also drain workers / close the
DB.
</details>

---

### 5. Why must the shutdown deadline be shorter than the orchestrator grace period?

<details>
<summary>Answer</summary>

After SIGTERM the orchestrator waits `terminationGracePeriodSeconds` (default 30s
on K8s) then sends SIGKILL. If your drain deadline is ≥ that, you get SIGKILLed
mid-drain — a slow shutdown *and* dropped requests. Keep it comfortably shorter
(e.g. 15s) so you finish draining before the hard kill.
</details>

---

### 6. Why flip readiness to "not ready" before draining?

<details>
<summary>Answer</summary>

So the load balancer stops routing *new* traffic a beat before you stop accepting
connections. If you drain while still "ready", new requests keep arriving into a
closing server. Readiness-off first, then drain, gives true zero-downtime
deploys.
</details>

---

### 7. What happens with `http.Server{}` and no timeouts?

<details>
<summary>Answer</summary>

The zero-value server has no `ReadTimeout`/`WriteTimeout`/`IdleTimeout`/
`ReadHeaderTimeout`. A slow client (slow-loris) can hold connections open
indefinitely by trickling bytes, exhausting connections/goroutines until the
server can't accept new work. Always set the timeouts explicitly.
</details>

---

### 8. (Extra) `Allow()` vs `Wait()` on a `rate.Limiter`?

<details>
<summary>Answer</summary>

`Allow()` is non-blocking: it returns false immediately if no token is available
(you respond 429). `Wait(ctx)` blocks until a token is free or `ctx` is cancelled
— useful for smoothing internal/outbound calls, not for rejecting abusive inbound
clients (where you want fast 429s, not held connections).
</details>

---

### 9. (Extra) How do you bound memory in a per-IP limiter?

<details>
<summary>Answer</summary>

The naive `map[string]*rate.Limiter` grows forever — one entry per IP, never
removed = a memory leak. Track last-seen time per entry and sweep idle ones
periodically, or use a TTL/LRU cache with a size cap. (Day 25 pitfall #7.)
</details>

---

### 10. (Extra) What's the correct shutdown ordering for a full service?

<details>
<summary>Answer</summary>

Stop intake → drain → close resources: (1) flip readiness off so the LB stops
new traffic, (2) `srv.Shutdown(ctx)` to drain in-flight HTTP, (3) cancel
background worker contexts and `wg.Wait()`, (4) close the DB pool, flush logs and
traces — all within the orchestrator grace period.
</details>

---

### 11. (Extra) Why pair server-side limits with client-side backoff?

<details>
<summary>Answer</summary>

They're two halves of overload control. Server limits shed load (429); client
backoff + jitter prevents synchronized retry storms that would otherwise re-create
the overload the instant capacity frees up. One without the other still lets a
retry storm take you down.
</details>

# Day 25 Pitfalls — Rate Limiting & Graceful Shutdown

Format: **Trap → Why → Fix**

---

### 1. `srv.Close()` instead of `srv.Shutdown(ctx)`
**Trap:** On SIGTERM you call `srv.Close()`.
**Why:** `Close()` immediately severs all active connections — in-flight requests get EOF/502. Every deploy produces a blip of 5xx that's impossible to reproduce locally.
**Fix:** `srv.Shutdown(ctx)` stops accepting new connections and drains in-flight requests within a deadline. (This is the Day 25 debugging exercise.)

---

### 2. Not waiting for `Shutdown` to return
**Trap:** Call `srv.Shutdown(ctx)` then immediately `os.Exit` / let `main` return.
**Why:** Exiting kills the drain you just started — the requests you meant to protect die anyway.
**Fix:** Block until `Shutdown` returns (and check its error). Only then exit.

---

### 3. Shutdown deadline ≥ orchestrator grace period
**Trap:** 30s shutdown timeout when K8s `terminationGracePeriodSeconds` is 30s (the default).
**Why:** If your drain runs to (or past) the grace period, you get SIGKILLed mid-drain — the worst of both worlds: a slow shutdown *and* dropped requests.
**Fix:** Make the shutdown deadline comfortably **shorter** than the grace period (e.g. 15s vs 30s).

---

### 4. `http.Server{}` with no timeouts
**Trap:** Using the zero-value `http.Server` for production.
**Why:** It has **no** `ReadTimeout`/`WriteTimeout`/`IdleTimeout`/`ReadHeaderTimeout`. A slow-loris client trickles bytes and ties up connections forever → resource exhaustion. This is a recurring real-world Go incident.
**Fix:** Set all four timeouts explicitly. Add `http.MaxBytesReader` for body limits and `MaxHeaderBytes`.

---

### 5. Assuming an in-memory limiter is cluster-wide
**Trap:** `rate.NewLimiter(10, 20)` and believing the service is limited to 10 rps.
**Why:** Each of N replicas has its own bucket, so the real limit is N× what you set. Behind a load balancer this silently multiplies your intended ceiling.
**Fix:** Know which you've built. For a true global limit, the counter must live in a shared store (Redis). In-memory is per-pod fairness, not a global cap.

---

### 6. Rate limiting without 429 + `Retry-After`
**Trap:** Rejecting with 500, or 429 but no headers.
**Why:** Clients can't tell they should back off, or for how long — they hammer harder and amplify the overload.
**Fix:** Return **429 Too Many Requests** with `Retry-After` and `X-RateLimit-Remaining`. Pair server limits with client-side backoff + jitter.

---

### 7. Per-IP limiter map that never evicts
**Trap:** `map[string]*rate.Limiter` that only grows.
**Why:** Every new client IP adds an entry that's never removed → unbounded memory growth → a slow leak that OOMs the pod.
**Fix:** Evict idle entries (track last-seen, sweep periodically) or use an LRU/TTL cache. Cap the map size.

---

### 8. Forgetting to drain workers / close the DB on shutdown
**Trap:** Graceful HTTP shutdown but background workers and the DB pool are abandoned.
**Why:** In-flight jobs are lost and connections leak; the shutdown isn't actually clean.
**Fix:** Order matters — stop intake (readiness off) → drain HTTP (`Shutdown`) → cancel worker contexts and `wg.Wait()` → close the DB pool / flush logs & traces.

# Day 29 — Pitfalls (performance & job queues)

Format: **Trap → Why → Fix**.

### 1. Optimizing without profiling
**Trap:** You rewrite a "slow-looking" loop and it makes no measurable difference (or hurts).
**Why:** Intuition about hot spots is usually wrong; the real cost is elsewhere.
**Fix:** Profile first (`pprof` CPU + alloc) under realistic load, fix what the profile points at, re-benchmark (`benchstat`).

### 2. `sync.Pool` for the wrong objects
**Trap:** Pooling long-lived or rarely-allocated objects adds complexity and bugs without speeding anything up.
**Why:** `sync.Pool` only pays off for short-lived objects allocated *millions* of times on a hot path.
**Fix:** Pool only what the alloc profile shows as hot and short-lived. Otherwise leave it alone.

### 3. Forgetting to `Reset()` a pooled object
**Trap:** A pooled buffer carries the previous user's bytes into the next call — data leaks between requests.
**Why:** `Get` returns a *reused* object that still holds old state.
**Fix:** `Reset()` (or zero) the object right after `Get`. And copy out anything the caller keeps — the object goes back to the pool.

### 4. Retries with no backoff/jitter
**Trap:** On a downstream blip, all clients retry immediately and in lockstep, amplifying the outage (retry storm).
**Why:** Fixed/instant retries synchronize and multiply load exactly when the system is weakest.
**Fix:** Exponential backoff (`base * 2^attempt`) **plus random jitter** so retries spread out. Cap the delay.

### 5. No dead-letter path
**Trap:** A permanently-failing "poison" job is retried forever, looping and starving good jobs.
**Why:** Without a terminal state, `maxRetries` is never enforced and the job never leaves the queue.
**Fix:** After `maxRetries`, move the job to a dead-letter collection for inspection; stop retrying it.

### 6. Unbounded job intake
**Trap:** A fast producer floods an unbounded queue; memory climbs until OOM.
**Why:** A slice-backed queue (or unbuffered hand-off with no limit) has no ceiling and no backpressure.
**Fix:** Bound the queue (a buffered channel of capacity N). Full → block the producer or shed load; never grow without limit.

### 7. Non-idempotent job handlers
**Trap:** A retried or redelivered job double-applies its effect (double email, double charge).
**Why:** At-least-once execution means a job can run more than once.
**Fix:** Make handlers idempotent — dedupe on a job ID, or use naturally idempotent operations / upserts.

### 8. Tuning `GOGC`/`GOMEMLIMIT` blindly
**Trap:** Someone sets `GOGC=400` or a memory limit "for performance" and OOMs or thrashes GC instead.
**Why:** These are trade-offs (memory vs GC frequency); the right value depends on the workload and container limits.
**Fix:** Change them only from data (profiles/observability). In containers, set `GOMEMLIMIT` to a soft ceiling below the cgroup limit to avoid OOM-kills.

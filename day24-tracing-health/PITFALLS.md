# Day 24 Pitfalls — Tracing & Health Checks

Format: **Trap → Why → Fix**

---

### 1. Dependency checks in liveness
**Trap:** `/healthz` (liveness) pings the database.
**Why:** A DB blip fails liveness → the orchestrator **restarts** every pod at once → they all reconnect together (thundering herd) → the DB dies for good → CrashLoopBackOff. A transient hiccup becomes a full outage.
**Fix:** Liveness = dumb 200 ("is the process alive?"). Put dependency checks only in **readiness** (`/readyz`), which pulls the pod from rotation without restarting it. (This is the Day 24 debugging exercise.)

---

### 2. Readiness that's too strict
**Trap:** Readiness fails if *any* dependency — including a non-critical one — is down.
**Why:** A flaky optional dependency flaps the pod in and out of rotation, causing instability worse than the dependency outage itself.
**Fix:** In readiness check only *critical* dependencies (the ones without which you literally can't serve). Degrade gracefully on optional ones.

---

### 3. Not propagating `ctx` → orphan spans
**Trap:** A function starts a span but the caller passed `context.Background()` instead of the request `ctx`.
**Why:** Child spans attach to the parent *through the context*. Break the chain and you get disconnected, orphan spans — the trace waterfall falls apart.
**Fix:** Thread the request `ctx` through every call. `ctx, span := tracer.Start(ctx, name)` and pass the *returned* ctx down.

---

### 4. Spans without `defer span.End()`
**Trap:** You `tracer.Start` but forget to `End()` on every return path.
**Why:** Unfinished spans leak and never appear (or appear with wrong durations); error returns silently skip the end.
**Fix:** `ctx, span := tracer.Start(ctx, name); defer span.End()` immediately, so every path closes the span.

---

### 5. Not recording errors on spans
**Trap:** A handler returns an error but the span isn't marked.
**Why:** The trace looks green while the request failed — you can't find failures by filtering for error spans.
**Fix:** `span.RecordError(err)` and `span.SetStatus(codes.Error, msg)` on failure paths.

---

### 6. Tracing 100% of requests at scale
**Trap:** No sampling — every request emits a full trace.
**Why:** Trace volume and export cost explode; you store mountains of boring successful traces.
**Fix:** Sample (e.g. 1–10%), or use **tail-based sampling** to keep all error/slow traces and a fraction of the rest. Use batched exporters to keep export off the hot path.

---

### 7. Health endpoints behind failing middleware/auth
**Trap:** `/healthz` and `/readyz` go through the same auth or rate-limit middleware as real traffic.
**Why:** If that middleware (or its dependency) fails, the probe fails for the wrong reason — the orchestrator reacts to a non-issue.
**Fix:** Register probe routes *before* heavy middleware, or exempt them. Keep probes fast, cheap, and dependency-light (cache heavy readiness checks briefly).

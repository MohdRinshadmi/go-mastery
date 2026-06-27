# Day 24 Interview Questions — Tracing & Health Checks

Lesson questions plus extras. Answers in `<details>`.

---

### 1. Trace vs span vs trace ID — define each.

<details>
<summary>Answer</summary>

A **trace** is the whole journey of one request across services. A **span** is
one unit of work within it (an HTTP handler, a DB query) with start/end time, a
parent, and attributes. A **trace ID** is the identifier shared by every span in
the same trace, so a tool can stitch them into one waterfall.
</details>

---

### 2. How does a trace span multiple services?

<details>
<summary>Answer</summary>

**Context propagation.** The trace ID and parent span ID travel between services
in HTTP headers — the W3C `traceparent` header. The receiving service reads them
and starts its spans as children of the caller's span, so spans in different
processes join one trace. `otelhttp`/`otelgrpc` do this automatically.
</details>

---

### 3. Why does every function taking `ctx` matter for tracing?

<details>
<summary>Answer</summary>

The current span lives in the `context.Context`. `tracer.Start(ctx, ...)` reads
the parent span from `ctx` and returns a new `ctx` carrying the child. If you
pass `ctx` down, child spans auto-attach to their parent; if you drop it (pass
`context.Background()`), spans become orphans and the trace breaks. It's the
deepest reason for the ctx-first convention.
</details>

---

### 4. Liveness vs readiness — what does each control, and the danger of conflating them?

<details>
<summary>Answer</summary>

**Liveness** (`/healthz`): "is the process alive?" — failing it **restarts** the
pod. **Readiness** (`/readyz`): "can I serve now?" — failing it **stops traffic**
(no restart). Conflating them — putting a DB check in liveness — turns a
transient dependency blip into a fleet-wide restart storm: liveness fails, all
pods restart and reconnect at once, and the dependency dies for good.
</details>

---

### 5. Why must liveness NOT check the database?

<details>
<summary>Answer</summary>

Because liveness failure triggers a restart. A DB hiccup would restart every pod
simultaneously; they all reconnect in a thundering herd that finishes off the DB,
producing a restart loop. A DB outage should pull pods from rotation (readiness),
not kill the processes. Liveness must depend only on the process itself.
</details>

---

### 6. Why sample traces, and what is tail-based sampling?

<details>
<summary>Answer</summary>

At scale, tracing every request is expensive in CPU, network, and storage — and
most traces are boring successes. Sampling keeps a fraction. **Head-based**
decides at the start (e.g. 5% randomly). **Tail-based** waits until the trace
finishes and keeps the interesting ones — all errors and slow traces — plus a
sample of the rest, so you never lose the traces you actually need.
</details>

---

### 7. How do you correlate a log line with its trace?

<details>
<summary>Answer</summary>

Put the trace ID (and span ID) into your structured logs:
`slog.With("trace_id", traceID)`. Then a log line links to its trace and the
trace's span links back to its logs. Correlating logs, metrics, and traces by
shared IDs is what "observability" means in practice.
</details>

---

### 8. (Extra) What is the W3C Trace Context standard?

<details>
<summary>Answer</summary>

A spec defining standard HTTP headers — primarily `traceparent` (version, trace
ID, parent span ID, flags) and `tracestate` — for propagating trace context
across services regardless of language or vendor. It's why polyglot services and
different tracing backends can share one trace.
</details>

---

### 9. (Extra) Should health endpoints sit behind your auth middleware?

<details>
<summary>Answer</summary>

No. If probes traverse auth/rate-limit middleware and that middleware (or its
dependency) fails, the probe fails for an unrelated reason and the orchestrator
restarts or de-rotates the pod wrongly. Register probe routes before heavy
middleware, or explicitly exempt them, and keep them cheap.
</details>

---

### 10. (Extra) Why `defer span.End()` and `span.RecordError`?

<details>
<summary>Answer</summary>

`defer span.End()` guarantees the span closes on every return path, including
errors — otherwise spans leak or get wrong durations. `span.RecordError(err)` +
`span.SetStatus(codes.Error, ...)` mark failures so the trace shows red where it
broke; without them a failing request can look perfectly healthy in the trace UI.
</details>

---

### 11. (Extra) How should readiness behave during a graceful shutdown?

<details>
<summary>Answer</summary>

Flip readiness to "not ready" *first*, so the load balancer stops sending new
traffic, then begin draining in-flight requests. This readiness-then-drain dance
(with Day 25's `srv.Shutdown`) is what makes deploys truly zero-downtime.
</details>

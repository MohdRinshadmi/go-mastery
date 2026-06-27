# Day 24 — Distributed Tracing (OpenTelemetry) & Health Checks

> Mentor note: Metrics (Day 23) tell you *that* p99 latency spiked. They don't tell you *where* the time went when a single request touches 6 services. That's **distributed tracing**: every request carries a trace ID, each service records spans, and a tool (Jaeger/Tempo) stitches them into one waterfall so you can see "the 800ms was the inventory service's DB call." Plus today: **health checks** — the tiny endpoints that let Kubernetes and load balancers know whether to send you traffic or restart you. Small, unglamorous, and the difference between a clean deploy and an outage.

---

## 1. Distributed tracing concepts

- **Trace** — the whole journey of one request across services. Identified by a **trace ID**.
- **Span** — one unit of work within a trace (an HTTP handler, a DB query, an RPC call). Has a start/end time, a parent span, and attributes.
- **Context propagation** — the trace ID + parent span ID travel between services in HTTP headers (`traceparent`, the W3C Trace Context standard). This is how spans in different processes join one trace.

A trace is a tree of spans you visualize as a waterfall:
```
[gateway 820ms]
  ├─[auth 12ms]
  ├─[catalog 40ms]
  └─[orders 760ms]
       └─[postgres query 740ms]   ← there's your culprit
```

## 2. OpenTelemetry (OTel) — the standard

OpenTelemetry is the vendor-neutral standard (CNCF) for traces/metrics/logs. You instrument once with OTel APIs and export to any backend (Jaeger, Tempo, Datadog, Honeycomb).

`go get go.opentelemetry.io/otel go.opentelemetry.io/otel/sdk go.opentelemetry.io/otel/exporters/...`

```go
// set up a tracer provider with an exporter (stdout/OTLP) once at startup
tp := trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(res))
otel.SetTracerProvider(tp)
defer tp.Shutdown(ctx)

tracer := otel.Tracer("orders-service")

func PlaceOrder(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "PlaceOrder")  // start a span
    defer span.End()                               // end it on return
    span.SetAttributes(attribute.String("user.id", uid))

    if err := charge(ctx); err != nil {            // ctx carries the span
        span.RecordError(err)
        span.SetStatus(codes.Error, "charge failed")
        return err
    }
    return nil
}
```

Key points:
- **Spans live in `context.Context`** — pass `ctx` down (Day 13!) and child spans automatically attach to the parent. This is the deepest reason every function takes `ctx`.
- **HTTP/gRPC instrumentation** (`otelhttp`, `otelgrpc`) auto-creates spans and propagates headers — you wrap your handler/client once and cross-service traces just work.
- **Sampling**: tracing every request is expensive at scale; sample a percentage (e.g. 1–10%) or use tail-based sampling to keep all the slow/error traces.

**Senior take:** The trace ID is the thread that ties everything together. Put it in your **logs** (Day 18) too — `slog.With("trace_id", traceID)` — so a log line links to its trace and vice versa. Logs + metrics + traces correlated by IDs is what "observability" actually means in practice.

## 3. Health checks

Two distinct endpoints (Kubernetes formalizes the distinction):

- **Liveness** (`/healthz`) — "is the process alive / not deadlocked?" If it fails, the orchestrator **restarts** the pod. Keep it dumb: return 200 if the process can serve. Do NOT check dependencies here — a DB blip shouldn't restart-loop your app.
- **Readiness** (`/readyz`) — "can I serve traffic *right now*?" Checks critical dependencies (DB reachable, caches warm, migrations done). If it fails, the orchestrator **stops sending traffic** but does NOT restart. Used during startup and transient dependency outages.

```go
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK) // alive
})

mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {   // dependency check
        http.Error(w, "db not ready", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

**Senior take:** The classic outage: putting a DB check in **liveness**. The DB hiccups → liveness fails → every pod restarts simultaneously → thundering herd reconnect → DB fully dies → restart loop. Liveness = "am I alive", readiness = "can I serve". Never check downstreams in liveness.

## Common mistakes
1. DB/dependency checks in liveness → restart storms.
2. Readiness that's too strict (fails on a non-critical dependency) → service flaps in and out of rotation.
3. Not propagating `ctx` → broken/disconnected traces (orphan spans).
4. Tracing 100% at scale → huge cost; no sampling strategy.
5. Spans without `defer span.End()` → leaked/unfinished spans.
6. Not recording errors on spans (`span.RecordError`) → traces look healthy while failing.
7. Health endpoints behind auth/middleware that can itself fail → probes break.

## Performance
- Span creation is cheap but not free; sample at scale. Batched exporters keep export off the hot path.
- Health checks must be fast and cheap (they're hit every few seconds per pod). Cache readiness results briefly if the dependency check is heavy.

---

## Expert Thinking Mode — "the request is slow somewhere"

- **Beginner:** "Add log lines everywhere and grep."
- **Senior:** "Trace it. Spans across services show the waterfall; the slow span is obvious. Trace ID in logs to drill in."
- **Staff:** "Sampling strategy (tail-based to keep errors/slow), context propagation across HTTP/gRPC/queues, span attributes that aid debugging without high cardinality. Health checks split correctly for safe rollout."
- **Architect:** "Tracing is a cross-team contract: consistent service names, propagation across every hop incl. async (Kafka), and a cost model for retention/sampling. Probes integrate with the deploy/rollout and autoscaling strategy."

---

## Real-world use

- **OpenTelemetry** is the industry standard; backends are Jaeger/Tempo/Datadog/Honeycomb. Uber's Jaeger originated this space.
- **W3C Trace Context** (`traceparent` header) propagates traces across polyglot services.
- **Liveness/readiness split** is baked into every Kubernetes deployment; the liveness-checks-DB anti-pattern is a famous incident category.

---

## Interview Questions

1. Trace vs span vs trace ID — define each.
2. How does a trace span multiple services? (context propagation / `traceparent`.)
3. Why does every function taking `ctx` matter for tracing?
4. Liveness vs readiness — what does each control, and what's the danger of conflating them?
5. Why must liveness NOT check the database?
6. Why sample traces, and what's tail-based sampling?
7. How do you correlate a log line with its trace?

---

## Your tasks

`../exercises/` has a service with health endpoints to implement: (1) `/healthz` liveness (dumb 200), (2) `/readyz` readiness that checks a simulated dependency with a timeout, and (3) a small manual "span" timer + a challenge to thread a trace ID through context and into structured logs. The runnable demo uses a stdlib-only mini-tracer; the real OTel setup is in `solutions/otel_reference.go` (build-ignored) with run instructions. Reference in `../solutions/`.

---

## Day 24 companion files

Self-contained study material for this day (in the day folder root):

- [Debugging exercise](../debugging/README.md) — liveness that checks the DB (the restart-storm anti-pattern); fixed by splitting liveness vs readiness ([bugged](../debugging/bugged/main.go) vs [fixed](../debugging/fixed/main.go)).
- [PITFALLS.md](../PITFALLS.md) — 7 tracing/health traps as Trap → Why → Fix.
- [INTERVIEW.md](../INTERVIEW.md) — interview Q&A with model answers.
- [NOTES.md](../NOTES.md) — quick reference + key terms.
- [RESOURCES.md](../RESOURCES.md) — curated links for Day 24.

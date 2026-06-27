# Day 24 Notes — Tracing & Health Checks (quick reference)

## Tracing vocabulary
- **Trace** — one request's whole journey across services (one trace ID).
- **Span** — one unit of work (handler, DB query); has parent, time, attributes.
- **Context propagation** — trace ID + parent span travel in the `traceparent`
  HTTP header (W3C Trace Context).

## Trace = tree of spans (waterfall)
```
[gateway 820ms]
  ├─[auth 12ms]
  ├─[catalog 40ms]
  └─[orders 760ms]
       └─[postgres 740ms]   ← the culprit
```

## OTel span pattern
```go
ctx, span := tracer.Start(ctx, "PlaceOrder") // reads parent from ctx
defer span.End()                             // close on every path
span.SetAttributes(attribute.String("user.id", uid))
if err := charge(ctx); err != nil {          // pass ctx DOWN
    span.RecordError(err)
    span.SetStatus(codes.Error, "charge failed")
    return err
}
```
Rules: pass `ctx` everywhere · `defer span.End()` · record errors · sample at scale.

## Health checks — the critical split
| Probe | Question | Fail action | Checks deps? |
|---|---|---|---|
| Liveness `/healthz` | "process alive?" | **restart** pod | **NO** |
| Readiness `/readyz` | "can I serve now?" | stop traffic (no restart) | yes (critical only) |

```go
mux.HandleFunc("/healthz", func(w, r) { w.WriteHeader(200) }) // dumb
mux.HandleFunc("/readyz", func(w, r) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        http.Error(w, "db not ready", 503); return
    }
    w.WriteHeader(200)
})
```

## The famous outage
DB check in **liveness** → DB blip → all pods restart at once → reconnect
stampede → DB dies → CrashLoopBackOff. Liveness = "am I alive", readiness =
"can I serve". Never check downstreams in liveness.

## Sampling
- Head-based: decide at start (e.g. 5%).
- Tail-based: keep all error/slow traces + a sample of the rest.
- Batched exporters keep export off the hot path.

## Key terms
- **traceparent** — W3C header carrying trace context.
- **Orphan span** — span detached from its parent (broken ctx propagation).
- **Liveness / Readiness** — restart vs de-rotate probes.
- **Thundering herd** — simultaneous reconnect storm after mass restart.
- **Tail-based sampling** — keep interesting traces after they complete.

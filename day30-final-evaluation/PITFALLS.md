# Day 30 — Pitfalls (microservices platform synthesis)

Format: **Trap → Why → Fix**. These are the cross-cutting traps that bite when you
wire Days 26–29 into one system.

### 1. Goroutine leak from an abandoned result channel
**Trap:** A per-request goroutine sends its result on an unbuffered channel; after the caller times out, the late send blocks forever and leaks.
**Why:** A send on an unbuffered channel needs a receiver; once the caller walks away there is none.
**Fix:** Buffer the result channel (cap 1) so the send always succeeds, and/or thread a `ctx` so the goroutine aborts on timeout. Every goroutine needs a guaranteed exit.

### 2. Choosing the wrong communication style per interaction
**Trap:** Routing a query that needs an immediate answer through async events, or coupling a fire-and-forget side effect into a synchronous call path.
**Why:** Sync gives an immediate answer but couples availability; async decouples but is eventually consistent.
**Fix:** Sync gRPC for queries/operations needing an answer now (auth, get product, stock check); async events for fan-out/work that can happen later (OrderPlaced → payment/inventory/email).

### 3. The dual-write problem (DB committed, event lost)
**Trap:** Write the order to the DB, then publish the event; the publish fails and the side effects never happen.
**Why:** Two systems (DB + broker) with no shared transaction can partially succeed.
**Fix:** Outbox pattern — write the order row AND an event row in one DB transaction; a relay publishes from the outbox.

### 4. No idempotency at a layer that retries
**Trap:** HTTP retries, Kafka redelivery, or job retries double-apply an effect because one consumer wasn't idempotent.
**Why:** Idempotency is needed at *every* layer that can duplicate — they don't cover for each other.
**Fix:** Idempotency keys at the gateway, dedupe-on-event-ID in every Kafka consumer, idempotent job handlers. Make it a checklist per layer.

### 5. No deadline / context propagation across the call chain
**Trap:** One slow downstream gRPC call hangs a request, and the hang cascades up because nothing cancels.
**Why:** Without a propagated deadline, each hop waits indefinitely.
**Fix:** `context.WithTimeout` on every sync call; pass `ctx` down so a parent timeout cancels all children (Day 25/26).

### 6. Premature microservices (distributed monolith)
**Trap:** Splitting into services before there's a real scaling/ownership boundary — now you have network failures and distributed debugging with none of the benefits.
**Why:** Services that must deploy together and share data are a monolith with added latency and failure modes.
**Fix:** Start from a modular monolith (Day 20); extract a service only when a clear scaling/ownership seam demands it.

### 7. Losing trace context at a boundary
**Trap:** A request is traceable within a service but the trace breaks crossing gRPC or Kafka, so end-to-end debugging is impossible.
**Why:** Trace context must be explicitly propagated in gRPC metadata and event headers.
**Fix:** Propagate `trace_id`/span context across every gRPC call and into Kafka message headers; log it everywhere (Days 23–24).

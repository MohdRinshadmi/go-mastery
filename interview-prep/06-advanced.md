# Phase 6 — Advanced Go & Distributed Systems (Days 26–30)

gRPC & protobuf, messaging (Kafka/RabbitMQ), caching & distributed-systems fundamentals, performance & job queues, system synthesis. Self-quiz: answer aloud, then expand.

---

### 1. gRPC Unary vs Server Streaming — give a real use case for each. Why HTTP/2?

<details><summary>Answer</summary>

**Unary** = one request, one response (`GetOrder(id) → Order`) — the RPC analog of a normal function call. **Server streaming** = one request, a *stream* of responses (`WatchOrders(filter) → stream OrderEvent`) — use it for live feeds, large result sets you want to process incrementally, or progress updates. gRPC uses **HTTP/2** for **multiplexing** (many concurrent streams over one TCP connection without head-of-line blocking at the HTTP layer), binary framing, header compression (HPACK), and native bidirectional streaming — things HTTP/1.1 can't do, where each in-flight request needs its own connection.
</details>

---

### 2. gRPC vs REST — when do you choose each?

<details><summary>Answer</summary>

**gRPC** for **internal service-to-service** calls: a strict schema (protobuf) with codegen, compact binary payloads, HTTP/2 streaming, and low latency — great for a microservice mesh. **REST/JSON** for **public/browser-facing** APIs: human-readable, debuggable with curl, universally supported, cache-friendly via HTTP semantics, and no special client tooling. gRPC's friction is browser support (needs grpc-web/a proxy) and opacity on the wire; REST's cost is verbosity and weaker contracts. Common pattern: gRPC inside, REST at the edge (often via a gateway).
</details>

---

### 3. Explain protobuf field numbers. Why can you never change one once deployed? How do you add a field safely?

<details><summary>Answer</summary>

Protobuf serializes by **field number, not name** — the number (plus wire type) is the on-wire tag, and names exist only in the schema. So changing a deployed field's number means old data is parsed into the wrong field (or dropped) — **silent, corrupting incompatibility** across any peers still using the old number. To **evolve safely**: only **add new fields with new, never-reused numbers**, keep them **optional** (you cannot add a truly "required" field without breaking old clients that won't send it — instead add it optional and validate at the app layer), never reuse a removed field's number (`reserve` it), and never change existing numbers or types.
</details>

---

### 4. `codes.Unavailable` vs `codes.Internal` — why does it matter for retries? What are interceptors?

<details><summary>Answer</summary>

`codes.Unavailable` signals a **transient** failure (server down, overloaded, connection dropped) — it's **safe to retry**, and clients/proxies will. `codes.Internal` signals a **bug/unexpected** server-side error — retrying just repeats the failure, so it's **not retried** by default. Returning the wrong code either causes retry storms on un-retryable errors or fails to recover from transient ones. **Interceptors** are gRPC middleware — they wrap RPCs for cross-cutting concerns (auth, logging, metrics, recovery). A unary server interceptor's signature:

```go
func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
     handler grpc.UnaryHandler) (resp any, err error)
```
You do work, then call `handler(ctx, req)` to proceed.
</details>

---

### 5. How do you load-balance gRPC in production, and why is L4 round-robin insufficient?

<details><summary>Answer</summary>

gRPC runs many requests over **one long-lived HTTP/2 connection**, so an **L4 (connection-level) load balancer** pins all of a client's traffic to whichever backend it first connected to — connections, not requests, get balanced, so load skews badly and new backends sit idle. You need **request/L7-aware balancing**: client-side load balancing (gRPC's `round_robin` policy over a resolver that returns all backend addresses), a service mesh (Envoy/Linkerd) that understands HTTP/2 streams, or a proxy that balances per-request. The key insight is "balance streams, not connections."
</details>

---

### 6. When choose async messaging over a synchronous call?

<details><summary>Answer</summary>

Choose async messaging when the caller **doesn't need the result immediately** and you want **decoupling, buffering, and resilience**: fan-out to multiple consumers, smoothing traffic spikes (the queue absorbs bursts), surviving a consumer being down (messages wait), and letting producer and consumer scale and deploy independently. Use a **synchronous** call when you need an immediate answer to proceed (a read, a validation, a user-facing response). The tradeoff: async buys decoupling and resilience but costs you eventual consistency, harder debugging, and the need for idempotency and ordering care.
</details>

---

### 7. Kafka vs RabbitMQ — model differences and when to use each?

<details><summary>Answer</summary>

**Kafka** is a **distributed, partitioned, replayable log**: messages persist, consumers track their own offset, and you can re-read history. Ideal for **high-throughput event streaming, event sourcing, and multiple independent consumer groups** replaying the same stream. **RabbitMQ** is a **traditional broker** with exchanges/queues and smart routing; messages are typically **consumed and removed**, with rich routing topologies and per-message acks. Ideal for **task/work queues, complex routing, and RPC-style** messaging. Rule of thumb: Kafka for streams and replay; RabbitMQ for flexible routing and job distribution.
</details>

---

### 8. What are partitions and consumer groups, and how do they enable scaling + ordering?

<details><summary>Answer</summary>

A topic is split into **partitions**, the unit of parallelism and ordering. Kafka guarantees order **only within a partition**, and messages with the same **key** hash to the same partition — so per-key order is preserved while different keys parallelize. A **consumer group** is a set of consumers that share the work: each partition is assigned to **exactly one consumer in the group**, so you scale by adding consumers up to the partition count, and each message is processed once *per group*. More partitions = more parallelism but more overhead and weaker global ordering.
</details>

---

### 9. Explain at-least-once delivery and why it forces idempotent consumers. How do you make a consumer idempotent?

<details><summary>Answer</summary>

**At-least-once** means the broker guarantees a message is delivered, but **may deliver it more than once** — because acks can be lost, a consumer can crash after processing but before committing its offset, causing redelivery. (True **exactly-once** end-to-end across systems is effectively unavailable — claiming it in an interview is a red flag.) So consumers **must be idempotent**: processing a duplicate must have no additional effect. Two techniques: (1) **dedup by message/event ID** — track processed IDs (a set/table) and skip seen ones; (2) **make the operation naturally idempotent** — use `UPSERT`/set-state-to-X rather than increment, or guard with a unique constraint / idempotency key so a replay is a harmless no-op.
</details>

---

### 10. What is a dead-letter queue, and what is consumer lag?

<details><summary>Answer</summary>

A **dead-letter queue (DLQ)** is where messages go after they **repeatedly fail** processing (exhaust retries) or are un-parseable — so one poison message doesn't block the queue or loop forever, and you can inspect/replay it later. It turns "stuck consumer" into "isolated, observable failure." **Consumer lag** is how far behind a consumer group is — the gap between the latest produced offset and the consumer's committed offset. You **monitor lag** because growing lag means consumers can't keep up with producers (under-scaled, slow handler, or stuck) — it's the leading indicator of a backing-up pipeline, often before users notice.
</details>

---

### 11. What is a cache stampede and how do you prevent it (name 3 techniques)? What is cache penetration?

<details><summary>Answer</summary>

A **stampede (thundering herd)** is many concurrent requests all missing the same hot key at once (e.g., right after it expires) and **all hitting the DB simultaneously**, potentially melting it. Prevent it with: (1) **single-flight** — collapse concurrent misses for the same key so only **one** does the DB load and the rest wait for its result (Go's `golang.org/x/sync/singleflight`); (2) **a short lock/lease** per key in a distributed cache so one worker recomputes; (3) **early/probabilistic expiration** (refresh slightly before TTL) to avoid synchronized expiry. **Cache penetration** is repeated lookups for keys that **don't exist**, which always miss and always hit the DB — defend by **caching the negative result** ("not found") with a short TTL (or a bloom filter).
</details>

---

### 12. State the CAP theorem. Why is there no "CA" system in practice? Strong vs eventual consistency?

<details><summary>Answer</summary>

CAP: under a **network partition (P)**, a distributed system must choose between **Consistency** (every read sees the latest write, or errors) and **Availability** (every request gets a — possibly stale — response). You can't have both *during a partition*. There's **no CA system in practice** because partitions are inevitable in real networks — "no partitions" isn't a choice you get, so you're always really choosing **CP** (refuse/err to stay consistent) or **AP** (serve possibly-stale to stay available). **Strong** consistency: reads always reflect the latest write — simple to reason about, costly to scale (Postgres-ish). **Eventual**: replicas converge "eventually," reads may be briefly stale — scales well (Cassandra/Dynamo, most caches). Read-your-writes/causal are useful middle grounds.
</details>

---

### 13. What is an idempotency key and what problem does it solve?

<details><summary>Answer</summary>

An **idempotency key** is a client-supplied unique ID for one *logical* operation; the server records keys it has processed and returns the original result for any **retry with the same key** instead of executing again. It solves the **duplicate-execution** problem caused by retries, redelivery, and timeouts — e.g., a payment request that times out and is retried must not double-charge. This is exactly how Stripe's API prevents double charges. It's the request-level expression of the same idempotency principle that makes message consumers and job handlers safe under at-least-once delivery.
</details>

---

### 14. Stack vs heap in Go — what decides, and how do you see escape decisions? Name four ways to reduce allocations.

<details><summary>Answer</summary>

The **compiler's escape analysis** decides: if a value's lifetime can be proven to end with the function (doesn't outlive the frame), it's stack-allocated (cheap, auto-freed); if its address escapes (returned, stored in a heap structure, captured by a long-lived closure, or passed to something that may retain it), it goes to the **heap** (GC-managed). See decisions with `go build -gcflags='-m'`. Four ways to cut allocations: (1) **preallocate** slices/maps with known capacity (`make([]T, 0, n)`); (2) **reuse buffers** via `sync.Pool`; (3) **avoid interface boxing / unnecessary pointers** that force escape; (4) **pass/return values** instead of pointers when the type is small enough that copying beats heap allocation (and avoid `[]byte`↔`string` conversions in hot loops).
</details>

---

### 15. When is `sync.Pool` appropriate vs a mistake? What do `GOGC` and `GOMEMLIMIT` control?

<details><summary>Answer</summary>

`sync.Pool` is appropriate for **short-lived, frequently-reused, same-type** objects in a hot path (per-request buffers, encoders) when profiling shows real allocation/GC pressure — it amortizes allocations and you **must reset** objects on get/put. It's a mistake for long-lived or rarely-reused objects, as a general object cache (pool entries can be **GC'd at any time**, so you can't rely on retention), or when it adds complexity without a measured win. **`GOGC`** (default 100) sets the GC trigger as a percentage of heap growth — raising it trades memory for fewer GC cycles (lower CPU), lowering it does the reverse. **`GOMEMLIMIT`** sets a **soft memory ceiling**: the runtime GCs more aggressively as you approach it, which prevents OOM kills in memory-capped containers (use it alongside, not instead of, `GOGC`).
</details>

---

### 16. Design a job queue — what makes it production-grade (3 things)? Why backoff *and* jitter, and why must handlers be idempotent?

<details><summary>Answer</summary>

Three production-grade properties: (1) **bounded concurrency** (a worker pool, so load is capped and predictable); (2) **retries with exponential backoff + jitter** for transient failures; and (3) a **dead-letter** path for jobs that exhaust retries, so failures are isolated and observable rather than infinitely looping. You need **backoff** to stop hammering a struggling dependency (giving it time to recover) **and jitter** to **desynchronize** retries — without jitter, many failed jobs retry in lockstep and create a coordinated thundering herd that re-overloads the dependency. Handlers must be **idempotent** because a job may run more than once (retry after a partial success, redelivery, worker crash); a non-idempotent handler double-applies effects (double charge, duplicate email) under exactly the conditions retries are meant to handle.
</details>

---

### 17. When is synchronous gRPC vs asynchronous events the right call between services? What is the outbox pattern?

<details><summary>Answer</summary>

Use **synchronous gRPC** when the caller needs an **immediate, authoritative answer** to continue (query a balance, validate a user, fetch the catalog for a page render) — request/response coupling is acceptable and you want the result now. Use **asynchronous events** when you're **notifying** rather than asking — fan-out, decoupling, buffering, and tolerating downstream being slow or down (an order placed → notify payment, inventory, email independently). The **outbox pattern** solves **dual-write** atomicity: writing to your DB *and* publishing an event aren't a single transaction, so a crash between them loses the event (or emits a phantom). Instead you write the event into an **outbox table in the same DB transaction** as the business change; a separate relay polls the outbox and publishes to the broker, giving you **atomic, at-least-once** event emission consistent with the DB write.
</details>

---

### 18. Choreography vs orchestration — trade-offs? Where do sagas fit?

<details><summary>Answer</summary>

**Choreography**: services react to each other's events with no central brain — each does its part and emits the next event. It's loosely coupled and scales organically, but the overall workflow is **emergent and hard to see/debug**, and changing it means touching many services. **Orchestration**: a central coordinator explicitly drives the steps. It's easy to understand and modify in one place, but the orchestrator is a coupling point and potential bottleneck. **Sagas** manage a **distributed transaction** as a sequence of local transactions, each with a **compensating action** to undo it on failure — they can be implemented either way (choreographed via events or orchestrated). Sagas are preferred over **two-phase commit** at scale because 2PC's locking and coordinator dependency don't survive partitions or high throughput.
</details>

---

### 19. How do you trace one request across gRPC and Kafka boundaries? How does idempotency show up at every layer?

<details><summary>Answer</summary>

You **propagate trace context across every boundary**: for gRPC, inject the trace context (`traceparent`) into the call metadata and extract it on the server; for **Kafka**, put the trace context in **message headers** so the consumer can continue the same trace when it processes the message later — otherwise the async hop breaks the trace into disconnected pieces. The key is that the trace ID rides along through synchronous *and* asynchronous edges. **Idempotency recurs at every layer** of a real system: **HTTP** (idempotency keys so a retried request doesn't double-act), **Kafka/messaging** (dedup by event ID under at-least-once redelivery), and **job queues** (retried handlers must not double-apply). It's one principle — "applying twice equals applying once" — enforced wherever duplicates can occur, which in a distributed system is *everywhere*.
</details>

---

### 20. Why start with a modular monolith instead of microservices? Walk through placing an order end to end.

<details><summary>Answer</summary>

Start with a **modular monolith** because microservices add enormous operational tax — network calls, distributed transactions, partial failure, deploy/observability complexity — that's only worth paying once you have **proven scaling/team boundaries**. A monolith with clean internal module boundaries (clear interfaces, no cross-module reach-in) gives you most of the design benefit with in-process simplicity, and you **extract a service only when a real pressure demands it** (independent scaling, team ownership, isolation). Premature microservices is the classic over-engineering trap. **End-to-end order**: client `POST /orders` → auth middleware validates the JWT and role → handler validates input and calls the order service → service checks stock (synchronous catalog read, possibly cache-aside) and writes the order **plus an outbox event in one DB transaction** → returns `201` to the client → the outbox relay publishes `OrderPlaced` to Kafka → **idempotent** consumers fan out (payment charges with an idempotency key, inventory decrements, email sends), each deduping by event ID → trace context flows through gRPC/HTTP and Kafka headers so the whole journey is one trace, and metrics/logs carry the correlation ID. That single narrative ties Phases 1–6 together.
</details>

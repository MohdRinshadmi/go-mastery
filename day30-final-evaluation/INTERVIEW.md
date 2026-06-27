# Day 30 — Interview Q&A (microservices synthesis)

<details>
<summary><strong>1. When do you use synchronous gRPC vs asynchronous events between services?</strong></summary>

Use **synchronous gRPC** for queries and operations that need an immediate answer: authenticate a token, fetch a product, check stock right now. You get low latency and strongly-typed contracts, but you couple the caller to the callee's availability. Use **asynchronous events** for fan-out and work that can happen later: `OrderPlaced` → payment, inventory, email each react independently. You get decoupling, resilience, and scalability at the cost of eventual consistency. The skill is choosing per interaction — gateway→catalog is sync (the user waits for the page); order→side-effects is async (the order is accepted immediately, effects follow reliably but later).
</details>

<details>
<summary><strong>2. What is the outbox pattern and what problem does it solve?</strong></summary>

It solves the **dual-write** problem: a service that writes to its DB *and* publishes an event can succeed at one and fail at the other (DB committed, event lost). With the outbox, the service writes the business row **and** an event row in the **same DB transaction** — atomic. A separate relay (or CDC) reads the outbox table and publishes to the broker, marking rows sent. So the state change and the intent-to-publish commit together, and the relay guarantees the event eventually goes out (at-least-once, hence idempotent consumers).
</details>

<details>
<summary><strong>3. Choreography vs orchestration — trade-offs? Where do sagas fit?</strong></summary>

**Choreography**: services react to each other's events with no central coordinator. Loosely coupled and easy to extend, but the end-to-end flow is implicit and hard to trace. **Orchestration**: a central coordinator explicitly drives the steps via commands. The flow is explicit and observable, at the cost of a central component and tighter coupling. **Sagas** are how you do a multi-service transaction without distributed 2PC: a sequence of local transactions, each with a **compensating action** to undo it if a later step fails. A saga can be choreographed (events trigger the next step) or orchestrated (a coordinator runs it).
</details>

<details>
<summary><strong>4. How do you trace one request across gRPC and Kafka boundaries?</strong></summary>

Propagate a **trace context** (trace ID + span ID, e.g. W3C `traceparent`) explicitly at every hop: inject it into gRPC **metadata** on outbound calls and extract it on the server side; for events, put it in the **Kafka message headers** so the consumer can continue the trace. Each service starts a child span and emits structured logs tagged with the `trace_id`. With OpenTelemetry instrumentation doing the inject/extract, one request becomes a single distributed trace spanning the gateway, the gRPC calls, and the async consumers (Days 23–24).
</details>

<details>
<summary><strong>5. Why start with a modular monolith instead of microservices?</strong></summary>

Microservices add real, permanent costs: network failures, eventual consistency, distributed debugging, and operational overhead. You only want them when the coupling/scaling/ownership pain is real. A **modular monolith** (Day 20) gives you clean internal boundaries (bounded contexts, no shared internals) with none of the distributed-systems tax, and it's a single deploy. When a clear seam needs independent scaling or team ownership, you extract that module into a service along its existing boundary. Splitting too early gives you a **distributed monolith** — services that must deploy together — which is the worst of both worlds.
</details>

<details>
<summary><strong>6. How does idempotency show up at every layer?</strong></summary>

It's the same principle wherever duplicates can occur. **HTTP**: an idempotency key (client-supplied unique ID) lets the server dedupe retried requests (Stripe). **Kafka/messaging**: at-least-once delivery means redelivery, so consumers dedupe on the event ID or use naturally idempotent operations. **Job queues**: at-least-once execution means a job can run twice, so handlers dedupe on a job ID or upsert. Each layer must handle its own duplicates — they don't cover for each other. The unifying rule: design every operation to be safe to apply more than once.
</details>

<details>
<summary><strong>7. Walk through what happens, end to end, when a user places an order.</strong></summary>

The client sends a REST request to the **Gateway** (BFF). The gateway makes synchronous gRPC calls: **Auth** validates the token, **Catalog** returns product details (served from a cache-aside layer in front of its DB, with single-flight on hot keys), and it asks the **Order Service** to create the order. The Order Service writes the order row **and** an `OrderPlaced` event row in one DB transaction (outbox), and returns success immediately — the user's request completes fast. A relay publishes `OrderPlaced` to the Kafka `orders` topic. Three consumer groups react independently and idempotently: **Payment** charges the card (dedupe on event ID so a redelivery doesn't double-charge), **Inventory** decrements stock, **Email** sends a confirmation. Every gRPC call had a deadline; trace context flowed across gRPC and into the Kafka headers, so the whole flow is one trace. The system is eventually consistent — the order is accepted now, side effects complete reliably but slightly later.
</details>

<details>
<summary><strong>8. How do you prevent a goroutine leak in fan-out / async request handling?</strong></summary>

Every goroutine you start needs a guaranteed way to finish. The classic leak: a goroutine sends its result on an **unbuffered** channel, but the caller times out and stops receiving, so the late send blocks forever. Fix it with a **buffered channel (cap 1)** so the send always succeeds, and/or pass a `context` so the goroutine aborts the downstream work on timeout. Treat "fire and forget" as "fire and leak" unless there's a receiver, a buffer slot, or a `ctx.Done()` exit. Monitor `runtime.NumGoroutine()` or the goroutine pprof profile — a steadily climbing count is a leak.
</details>

<details>
<summary><strong>9. What are the key cross-cutting concerns in the platform, and which day each comes from?</strong></summary>

**Caching** (Day 28): catalog uses cache-aside + single-flight in front of its DB. **Idempotency** (Days 27/28): every Kafka consumer dedupes on event ID. **Observability** (Days 23–24): `/metrics`, propagated trace context across gRPC and Kafka, structured logs with `trace_id`. **Resilience** (Day 25): timeouts/context deadlines on every gRPC call, gateway rate limiting, graceful shutdown, retries with backoff. **Health** (Day 24): liveness/readiness probes per service for safe rollouts. **Deployment** (Phase 5): each service independently deployable via Docker + CI/CD with health-gated rollouts.
</details>

<details>
<summary><strong>10. What are the real costs of microservices, and when are they worth it?</strong></summary>

Costs: network calls fail and add latency; the system becomes eventually consistent across services; debugging spans many processes; and you take on serious operational overhead (deploys, service discovery, observability, on-call). They're worth it when the benefits outweigh that — primarily **independent team ownership and deployment**, and **independent scaling** of a hot component. Microservices are as much an *organizational* solution as a technical one. The senior move is to reach for them only when coupling/scaling/ownership pain is real, starting from a clean modular monolith that can split along its existing seams.
</details>

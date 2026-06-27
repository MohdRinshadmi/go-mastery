# Day 30 — Microservices Communication Platform & Final Evaluation

> Mentor note: You made it. Thirty days ago you were new to Go; today we design a microservices platform and then I assess you like the end of a real bootcamp — honestly. This day has two halves: (1) the architecture lesson that ties Phase 6 together (gRPC + messaging + caching into one coherent system), and (2) the **final evaluation**: a comprehensive exam plus a straight-talk readiness assessment — your level, the jobs you're ready for, what's still missing, and what to build next. Do the exam in `../exercises/FINAL_EXAM.md` before reading the answer key.

---

## Part 1 — Microservices Communication Platform

A realistic e-commerce platform decomposed into services, using everything from Phase 6:

```
                    ┌─────────────┐
   client ──REST──▶ │  Gateway    │ ──gRPC──▶ Auth Service
                    │  (BFF)      │ ──gRPC──▶ Catalog Service ──cache──▶ Redis
                    └──────┬──────┘ ──gRPC──▶ Order Service
                           │                       │
                           │                       └──emits event──▶ Kafka topic "orders"
                           │                                              │
                    (sync, low-latency)            ┌─────────────────────┼───────────────┐
                                                    ▼                     ▼               ▼
                                              Payment Svc          Inventory Svc     Email Svc
                                            (consumer group)     (consumer group)  (consumer group)
```

### The two communication styles, and when each
- **Synchronous (gRPC, Day 26)** — for *queries* and operations needing an immediate answer: "authenticate this token", "get this product", "is this in stock right now". Low latency, strongly typed contracts (protobuf), but couples caller to callee's availability.
- **Asynchronous (Kafka, Day 27)** — for *events / fan-out / work that can happen later*: `OrderPlaced` → payment + inventory + email each react independently. Decoupled, resilient, scalable; eventually consistent.

The art is choosing per interaction. Gateway→Catalog is sync (user waits for the page). Order→(payment/inventory/email) is async (the order is accepted immediately; side effects happen reliably but later).

### The cross-cutting concerns you now know how to handle
- **Caching (Day 28)**: Catalog uses cache-aside + singleflight in front of its DB.
- **Idempotency (Day 27/28)**: every Kafka consumer dedupes on event ID — payment must not double-charge on redelivery.
- **Observability (Days 23–24)**: each service exposes `/metrics`, propagates trace context across gRPC and Kafka, emits structured logs with `trace_id`. One request is traceable across the whole mesh.
- **Resilience (Day 25)**: timeouts + context deadlines on every gRPC call, rate limiting at the gateway, graceful shutdown everywhere, retries with backoff on transient failures.
- **Health (Day 24)**: liveness/readiness per service for safe rollouts.

### Key design decisions & trade-offs
- **Choreography vs orchestration**: events reacting to events (choreography — what we drew) is loosely coupled but the end-to-end flow is implicit/hard to trace. An orchestrator (a Saga coordinator) makes the flow explicit at the cost of a central coordinator. Use sagas with **compensating actions** for multi-service transactions (no distributed 2PC).
- **The outbox pattern**: Order Service writes the order row AND an event row in one DB transaction; a relay publishes to Kafka. Solves "DB committed but the event publish failed" — atomicity between state change and event.
- **Service boundaries = bounded contexts** (DDD): split by business capability (auth, catalog, orders), not by technical layer. Each owns its data; no shared database.
- **Modular monolith first**: the Day 20 layered monolith is often the *right* starting point. Extract a service only when a clear scaling/ownership boundary demands it. Premature microservices = distributed monolith = worst of both worlds.

### Production concerns (Staff/Architect view)
- **Scalability**: stateless services scale horizontally; Kafka partitions bound consumer parallelism; Redis offloads hot reads. The DB is usually the first bottleneck — cache and read-replicas.
- **Failure isolation**: bulkheads (a slow Payment service doesn't stall order intake because it's async), circuit breakers on sync calls, DLQs for poison events.
- **Consistency**: the system is eventually consistent across services; the UI/clients are designed to tolerate it (optimistic updates, status polling).
- **Deployment**: each service independently deployable (Phase 5: Docker, CI/CD, health-gated rollouts).

**Senior take:** Microservices are an *organizational* solution as much as a technical one — they let teams own and deploy independently. They add real cost: network failures, eventual consistency, distributed debugging, operational overhead. The senior move is to reach for them only when the coupling/scaling/ownership pain is real, and to start from a clean modular monolith (Day 20) that can split along its existing seams.

---

## Part 2 — FINAL EVALUATION

### How to do it
1. Complete the exam in [`../exercises/FINAL_EXAM.md`](../exercises/FINAL_EXAM.md): theory, coding, debugging, and code-review challenges spanning all 30 days.
2. Then check yourself against [`../solutions/ANSWER_KEY.md`](../solutions/ANSWER_KEY.md) and score with the rubric there.
3. Bring me your scored exam + any solution code and I'll give you the real assessment.

### The readiness framework I'll assess you on

**1. Your current level** — measured across: Go fundamentals, concurrency, backend/API, database/repository design, testing, observability, Docker/CI-CD, distributed systems, and performance. Rough bands:
- *Advanced Beginner* — writes correct Go, knows the idioms, builds a CRUD service.
- *Capable / Job-ready (Mid)* — clean architecture, solid concurrency under `-race`, tests, can operate a service (metrics/health/shutdown).
- *Senior-track* — designs systems, reasons about trade-offs (consistency, caching, messaging), debugs production, mentors on idioms.

**2. Jobs you're ready for** (after passing this program with real practice):
- **Go Backend Engineer** — REST/gRPC services, Postgres, Redis, Docker. ✅ core target.
- **Platform / Infra-adjacent** — if you lean into Phase 5 (observability, CI/CD, K8s).
- **Distributed Systems / Microservices Engineer** — with more depth on Phase 6 (Kafka at scale, real gRPC, sagas).
*Note:* this program makes you *interview-capable and productive*; seniority comes from shipping real systems and incidents.

**3. Skills likely still missing** (be honest with yourself):
- Production Kubernetes (we covered concepts, not deep ops).
- Real protobuf codegen pipelines and gRPC streaming at scale.
- Deep SQL (query optimization, transactions/isolation levels, migrations under load).
- Security depth (authz models, secrets management, supply chain).
- Running real incidents — the thing only a job gives you.

**4. Resume projects to add** (build these to prove the skills):
- The **Day 20 e-commerce API**, fleshed out with real Postgres + Redis + JWT + tests + Docker + a CI pipeline. One polished, complete service beats five toys.
- A **distributed job queue** (Day 29) backed by Redis, deployed with metrics + graceful shutdown.
- An **event-driven service** with Kafka, idempotent consumers, DLQ, and traces.
- Each with a README, tests, `docker-compose`, and a CI badge — show you can *operate*, not just code.

**5. What to learn next** (your 60-day path):
- Ship one of the above to a real cloud (Fly.io/Render/GCP) with CI/CD and observability.
- Deepen SQL + a real ORM/sqlc; learn transactions and migrations cold.
- Learn Kubernetes basics (deployments, probes, HPA) — Phase 5 made you ready for it.
- Read others' Go: the standard library, then a production codebase. Contribute a small OSS PR.
- Practice system design interviews using the trade-offs from Phases 4–6.

---

## Interview Questions (the synthesis set)
1. When do you use synchronous gRPC vs asynchronous events between services?
2. What is the outbox pattern and what problem does it solve?
3. Choreography vs orchestration — trade-offs? Where do sagas fit?
4. How do you trace one request across gRPC and Kafka boundaries?
5. Why start with a modular monolith instead of microservices?
6. How does idempotency show up at every layer (HTTP retries, Kafka redelivery, job retries)?
7. Walk me through what happens — end to end — when a user places an order in the platform above.

---

## The runnable demo (this folder)

`examples/` simulates the platform **in one process**: an in-memory event bus (Kafka stand-in), a cached catalog (cache-aside), an order "service" that emits `OrderPlaced`, and three idempotent consumers (payment, inventory, email). It ties Days 26–29 together so you can watch a synchronous catalog read + an asynchronous order fan-out in one run. Run: `go run .`

Congratulations on reaching Day 30. Do the final exam — then let's talk about where you go next. 🎓

# Day 30 — Quick Reference (microservices platform synthesis)

## The platform (Phase 6 tied together)
```
client ──REST──▶ Gateway (BFF) ──gRPC──▶ Auth
                              ──gRPC──▶ Catalog ──cache-aside+singleflight──▶ Redis
                              ──gRPC──▶ Order ──outbox──▶ Kafka "orders"
                                                              │
                                  ┌───────────────┬───────────┴────────────┐
                                  ▼               ▼                        ▼
                            Payment Svc      Inventory Svc             Email Svc
                          (consumer group)  (consumer group)        (consumer group)
                           idempotent         idempotent              idempotent
```

## Two communication styles — choose per interaction
| | Sync (gRPC, Day 26) | Async (Kafka, Day 27) |
|---|---|---|
| For | queries / immediate answer | events / fan-out / deferrable work |
| Examples | auth, get product, stock check | OrderPlaced → payment/inventory/email |
| Pro | low latency, typed contract | decoupled, resilient, scalable |
| Con | couples caller to callee uptime | eventually consistent |

## Cross-cutting concerns (and their day)
- **Caching (28):** cache-aside + single-flight + TTL in front of catalog DB.
- **Idempotency (27/28):** dedupe at every retrying layer (HTTP key, event ID, job ID).
- **Observability (23–24):** /metrics, trace context across gRPC + Kafka, structured logs w/ trace_id.
- **Resilience (25):** deadlines on every gRPC call, gateway rate limit, graceful shutdown, backoff retries.
- **Health (24):** liveness/readiness per service for safe rollouts.

## Key patterns
- **Outbox** — write business row + event row in one DB tx; relay publishes. Solves dual-write.
- **Choreography vs orchestration** — events-react-to-events (loose, implicit flow) vs central coordinator (explicit, coupled). **Sagas** = multi-service txn via local txns + **compensating actions** (not 2PC).
- **Bounded contexts (DDD)** — split by business capability; each owns its data; no shared DB.
- **Modular monolith first** — extract a service only at a real scaling/ownership seam; premature split = distributed monolith.

## Production concerns (Staff/Architect)
- Scalability: stateless services scale horizontally; Kafka partitions bound consumer parallelism; Redis offloads hot reads; DB is usually the first bottleneck.
- Failure isolation: bulkheads (async payment doesn't stall order intake), circuit breakers on sync calls, DLQs for poison events.
- Consistency: eventually consistent across services; UI tolerates it (optimistic updates, status polling).

## Final evaluation
- Do `../exercises/FINAL_EXAM.md` first, then score against `../solutions/ANSWER_KEY.md`.
- Readiness bands: Advanced Beginner → Capable/Job-ready (Mid) → Senior-track.
- Resume projects: polished Day 20 e-commerce API, a Redis-backed job queue, an event-driven service w/ DLQ + traces.

## Key terms
**BFF/gateway** · **sync vs async** · **outbox** · **choreography vs orchestration** · **saga / compensating action** · **bounded context** · **modular monolith vs distributed monolith** · **idempotency (every layer)** · **deadline propagation** · **bulkhead / circuit breaker / DLQ** · **eventual consistency** · **goroutine leak**.

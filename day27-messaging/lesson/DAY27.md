# Day 27 — Messaging: Kafka, RabbitMQ & Event-Driven Architecture

> Mentor note: Up to now your services call each other synchronously: A calls B, waits, gets a reply. That couples them — if B is down, A fails; if B is slow, A is slow. **Event-driven architecture** flips this: A emits an *event* ("OrderPlaced") to a broker and moves on; whoever cares (email service, inventory, analytics) consumes it independently. This is how you build systems that scale and tolerate partial failure. Today: the two dominant brokers (Kafka, RabbitMQ), when to use each, and the hard parts (delivery guarantees, idempotency) that separate a demo from production.

---

## 1. Sync vs async — the core trade

**Synchronous (REST/gRPC):** request → wait → response. Simple, immediate consistency, easy to reason about. But tight coupling and cascading failures: if the payment service is slow, every order request is slow.

**Asynchronous (messaging):** producer emits an event → broker → consumers process later. Decoupled, resilient (broker buffers when a consumer is down), scalable (add consumers). Cost: eventual consistency and more moving parts. You give up "it happened" certainty for "it will happen" durability.

Use async when: work can happen later (emails, receipts, analytics), you need to fan out one event to many consumers, you need to absorb load spikes (the queue is a shock absorber), or you want services to deploy/fail independently.

## 2. Kafka vs RabbitMQ

| | **Kafka** | **RabbitMQ** |
|---|---|---|
| Model | Distributed **log** (append-only, retained) | **Message queue** / broker |
| Consumption | Consumers track an **offset**; messages stay after reading (replayable) | Messages **removed** once acked |
| Throughput | Very high (millions/sec), built for streams | High, but lower than Kafka |
| Ordering | Per-partition ordering | Per-queue, weaker guarantees |
| Best for | Event streaming, event sourcing, log/metrics pipelines, replay | Task queues, RPC, complex routing (topics/fanout/direct exchanges) |
| Mental model | "A durable, replayable river of events" | "A smart post office routing letters to mailboxes" |

**Pick Kafka** for high-volume event streams you may want to replay or have many independent consumer groups read (analytics + search-index + notifications all reading the same `orders` topic). **Pick RabbitMQ** for work queues, request/reply, and rich routing where messages are consumed once and discarded.

### Kafka core concepts
- **Topic** — a named stream (`orders`). Split into **partitions** for parallelism.
- **Partition** — ordered, append-only sequence. Ordering is guaranteed *within* a partition only. The **key** (e.g. `order_id`) decides the partition, so all events for one order are ordered.
- **Consumer group** — consumers sharing the work of a topic; each partition goes to one consumer in the group. Add consumers (up to #partitions) to scale. Different groups each get the *full* stream independently.
- **Offset** — a consumer's position. Committing the offset = "I've processed up to here."

## 3. Delivery guarantees — the part that bites

- **At-most-once** — may lose messages (commit offset before processing). Rarely what you want.
- **At-least-once** — never lose, but may **duplicate** (process, then commit; a crash between = reprocess). The common default.
- **Exactly-once** — hard and expensive; Kafka offers it in narrow cases (transactions), but in practice you achieve "effectively once" via **at-least-once + idempotent consumers**.

### Idempotency is non-negotiable
Because at-least-once means duplicates, **your consumer must handle the same event twice without double-effect.** Techniques:
- Dedup on an **event ID** (store processed IDs; skip seen ones).
- Make the operation naturally idempotent (`SET status=paid` not `INCREMENT attempts`).
- Upserts keyed by the business ID.

```go
func handleOrderPaid(ctx context.Context, evt OrderPaid) error {
    if alreadyProcessed(evt.EventID) {  // idempotency guard
        return nil                       // duplicate -> no-op
    }
    if err := markOrderPaid(ctx, evt.OrderID); err != nil {
        return err                       // return error -> message redelivered
    }
    return recordProcessed(evt.EventID)
}
```

**Senior take:** "We use a queue so we won't lose messages" is half the lesson. The other half: **you will get duplicates, so every consumer must be idempotent.** Design the consumer assuming each event arrives 1+ times and out of order. If you can't, you have a correctness bug waiting for the next broker rebalance.

## 4. Other production concerns
- **Dead-letter queue (DLQ)** — after N failed retries, route the message aside so it doesn't block the queue and can be inspected. A poison message without a DLQ stalls a partition forever.
- **Backpressure / lag** — monitor **consumer lag** (how far behind the latest offset). Growing lag = consumers can't keep up; scale them or shed load.
- **Schema evolution** — events are a contract over time. Use a schema registry (Avro/Protobuf) and make changes backward-compatible; old consumers must not break on new fields.
- **Ordering vs parallelism** — more partitions = more parallelism but ordering only within a partition. Key by the entity that needs ordering.

## Common mistakes
1. Non-idempotent consumers with at-least-once delivery → double charges, duplicate emails.
2. No DLQ → one poison message blocks a partition/queue.
3. Committing the offset before processing → lost messages on crash.
4. Ignoring consumer lag → silent unbounded backlog.
5. Assuming global ordering across partitions (there is none).
6. Breaking schema compatibility → consumers crash on deploy.
7. Using a queue for something that needs an immediate synchronous answer.

## Performance
- Kafka throughput comes from partitioning + batching + sequential disk writes; tune batch size and partition count.
- Consumer parallelism caps at partition count — size partitions for peak.
- Idempotency stores (dedup tables) must be fast; Redis/DB with a TTL is common.

---

## Expert Thinking Mode — "service A needs to notify service B"

- **Beginner:** "A calls B's HTTP endpoint and waits."
- **Senior:** "Does B need to answer now, or just eventually? If eventually, emit an event — decouple them. At-least-once + idempotent consumer + DLQ."
- **Staff:** "Kafka vs RabbitMQ by access pattern; partition key for ordering; consumer-group scaling; lag monitoring; schema registry; the consistency model (eventual) and how the UI/clients cope."
- **Architect:** "Event-driven is a system topology choice: choreography vs orchestration, the event taxonomy as an org-wide contract, outbox pattern for atomic DB-write+publish, and the failure/replay story. Events become the integration backbone."

---

## Real-world use

- **Uber/Netflix/LinkedIn** run massive Kafka pipelines (LinkedIn created Kafka) for events, metrics, and stream processing.
- **Order systems**: `OrderPlaced` fans out to payment, inventory, email, analytics — each an independent consumer group.
- **Outbox pattern**: write the DB row and an event row in one transaction, then a relay publishes — solving "DB committed but publish failed."
- **RabbitMQ** powers task queues and RPC in countless backends; Go clients: `segmentio/kafka-go`, `rabbitmq/amqp091-go`.

---

## Interview Questions

1. When do you choose async messaging over a synchronous call?
2. Kafka vs RabbitMQ — model differences and when to use each?
3. What are partitions and consumer groups, and how do they enable scaling + ordering?
4. Explain at-least-once delivery. Why does it force idempotent consumers?
5. How do you make a consumer idempotent? Give two techniques.
6. What is a dead-letter queue and what problem does it solve?
7. What is consumer lag and why monitor it?

---

## Your tasks

`../exercises/` has an in-memory event bus (a stand-in for Kafka, runs offline) and an **OrderPlaced → [inventory, email]** event flow to wire up — including an **idempotent** consumer that ignores duplicate event IDs (the bus deliberately delivers one event twice so you can prove your dedup works). Real Kafka and RabbitMQ producer/consumer code is in `solutions/broker_reference.go` (build-ignored) with a `docker-compose.yml` to run the brokers. Reference in `../solutions/`.

---

## Day 27 companion files

Self-study companions for this day (in `../`):

- [`debugging/`](../debugging/) — the double-charge bug (at-least-once delivery without idempotency) with `bugged/` and `fixed/`.
- [`PITFALLS.md`](../PITFALLS.md) — messaging gotchas as Trap → Why → Fix.
- [`INTERVIEW.md`](../INTERVIEW.md) — interview questions with model answers.
- [`NOTES.md`](../NOTES.md) — quick reference + key terms.
- [`RESOURCES.md`](../RESOURCES.md) — curated links (Kafka, RabbitMQ, outbox/saga patterns).

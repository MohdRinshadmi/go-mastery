# Day 27 — Interview Q&A (messaging / event-driven)

<details>
<summary><strong>1. When do you choose async messaging over a synchronous call?</strong></summary>

When the work can happen *later* (emails, receipts, analytics), when one event must fan out to many independent consumers, when you need to absorb load spikes (the queue is a shock absorber), or when you want services to deploy and fail independently. You trade immediate consistency ("it happened") for durability ("it will happen") and accept eventual consistency plus more moving parts. If the caller needs an immediate answer (a query, a synchronous validation), keep it sync.
</details>

<details>
<summary><strong>2. Kafka vs RabbitMQ — model differences and when to use each?</strong></summary>

Kafka is a distributed, append-only **log**: messages are retained and consumers track an **offset**, so the stream is replayable and multiple consumer groups can each read the full stream independently. RabbitMQ is a **message broker/queue**: messages are removed once acked, with rich routing (direct/topic/fanout exchanges). Use **Kafka** for high-volume event streams, event sourcing, and replay (analytics + search + notifications all reading one `orders` topic). Use **RabbitMQ** for task queues, request/reply, and complex routing where each message is consumed once and discarded.
</details>

<details>
<summary><strong>3. What are partitions and consumer groups, and how do they enable scaling and ordering?</strong></summary>

A topic is split into **partitions** — ordered, append-only sequences. Ordering is guaranteed only *within* a partition; the message **key** decides the partition, so all events for one key (e.g. an `order_id`) stay ordered. A **consumer group** shares the work of a topic: each partition is assigned to exactly one consumer in the group, so you scale by adding consumers up to the partition count. Different groups each get the full stream independently. So partitions give parallelism *and* per-key ordering; consumer groups give horizontal scaling.
</details>

<details>
<summary><strong>4. Explain at-least-once delivery. Why does it force idempotent consumers?</strong></summary>

At-least-once means the broker guarantees a message is delivered, but possibly more than once: if the consumer processes a message and crashes (or the ack is lost) before committing, the broker redelivers it. So duplicates are normal. If the consumer's side effect isn't idempotent, a duplicate causes a double-effect (double charge, double email). Therefore the consumer must produce the same result whether it sees an event once or many times — that's idempotency, and it's what turns at-least-once into "effectively once."
</details>

<details>
<summary><strong>5. How do you make a consumer idempotent? Give two techniques.</strong></summary>

(1) **Dedup on a unique event ID**: store processed IDs (Redis SET / DB unique key with a TTL); skip any ID you've already handled — making the check-and-record atomic. (2) **Make the operation naturally idempotent**: `SET status=paid` instead of `balance += amount`, or an **upsert** keyed by the business ID (`INSERT ... ON CONFLICT DO NOTHING`). Re-applying either changes nothing.
</details>

<details>
<summary><strong>6. What is a dead-letter queue and what problem does it solve?</strong></summary>

A DLQ is a separate destination where a message is routed after it fails processing N times. It solves the **poison message** problem: without a DLQ, a message that always fails gets redelivered forever, blocking the partition/queue behind it so nothing else makes progress. The DLQ moves the bad message aside so the stream continues, and lets you inspect/replay it later.
</details>

<details>
<summary><strong>7. What is consumer lag and why monitor it?</strong></summary>

Consumer lag is how far behind the latest produced offset a consumer group is — the size of the unprocessed backlog. It's the primary health signal for a consumer: steady or zero lag means it's keeping up; growing lag means consumers can't keep pace and the backlog (and end-to-end latency) is climbing. You respond by scaling consumers (up to the partition count), optimizing the handler, or shedding load. Silent unbounded lag is how async pipelines fail invisibly.
</details>

<details>
<summary><strong>8. What is the outbox pattern and what does it solve?</strong></summary>

It solves the "dual write" problem: a service that writes to its DB *and* publishes an event can succeed at one and fail at the other (DB committed, publish lost). With the outbox, the service writes the business row **and** an event row in the **same DB transaction** — atomic. A separate relay/CDC process reads the outbox table and publishes to the broker, marking rows sent. Now the state change and the intent-to-publish commit together; the relay guarantees the event eventually goes out (at-least-once, hence idempotent consumers).
</details>

<details>
<summary><strong>9. Choreography vs orchestration in event-driven systems?</strong></summary>

Choreography: services react to each other's events with no central coordinator (`OrderPlaced` → payment, inventory, email each react). Loosely coupled and easy to extend, but the end-to-end flow is implicit and hard to trace/reason about. Orchestration: a central coordinator (often a **Saga** orchestrator) explicitly drives the steps and issues commands. The flow is explicit and easier to monitor, at the cost of a central component and more coupling. Multi-service transactions use sagas with **compensating actions** rather than distributed 2PC.
</details>

<details>
<summary><strong>10. Why is there no real exactly-once delivery, and how do teams get "effectively once"?</strong></summary>

True exactly-once across a network is impossible in the general case because you can't atomically "deliver and ack" over an unreliable link — you must choose to risk loss (at-most-once) or duplication (at-least-once). Kafka offers exactly-once *semantics* in narrow, transactional cases (within Kafka, producer→topic→consumer), but for arbitrary side effects you achieve **effectively once** = at-least-once delivery + idempotent consumers. That combination is simpler, cheaper, and more robust than trying to make delivery itself exactly-once.
</details>

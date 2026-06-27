# Day 29 — Interview Q&A (performance & job queues)

<details>
<summary><strong>1. Stack vs heap in Go — what decides, and how do you see escape decisions?</strong></summary>

The compiler runs **escape analysis**: if it can prove a value doesn't outlive the function, it goes on the **stack** (free, auto-reclaimed on return). If the value might escape — returned by pointer, stored in an interface, captured by a closure that outlives the call, or its size/lifetime can't be bounded — it goes on the **heap** (GC-managed, future GC work). You see the decisions with `go build -gcflags='-m' ./...`, which prints "escapes to heap" / "does not escape" per allocation.
</details>

<details>
<summary><strong>2. Name four ways to reduce allocations.</strong></summary>

(1) **Pre-size** slices/maps: `make([]T, 0, n)` to avoid repeated regrowth. (2) Use **`strings.Builder`** instead of `+=` string concatenation in loops. (3) **Avoid `interface{}` boxing** in hot loops — generics keep types concrete and on the stack. (4) **Reuse buffers** with `sync.Pool` for short-lived, frequently-allocated objects. (Bonus: pass small structs by value to keep them on the stack; avoid `fmt` with `...interface{}` on hot paths.)
</details>

<details>
<summary><strong>3. When is `sync.Pool` appropriate, and when is it a mistake?</strong></summary>

Appropriate when a profile shows a **short-lived object allocated millions of times** on a hot path (buffers, encoders) and GC pressure is a measurable cost. A mistake when used as a default, or for **long-lived** objects (defeats the purpose) or **rarely-allocated** ones (no gain, added complexity). The two correctness traps: forgetting to `Reset()` pooled objects (data leaks between uses) and returning pooled memory to a caller (data race when it's reused). It's a cache, not a guarantee — entries can be cleared on GC.
</details>

<details>
<summary><strong>4. What do `GOGC` and `GOMEMLIMIT` control? When would you change them?</strong></summary>

`GOGC` (default 100) sets the heap-growth target: GC runs when the heap has grown `GOGC%` since the last collection. Higher (e.g. 200) = fewer GCs but more memory; lower = more GCs, less memory. `GOMEMLIMIT` (Go 1.19+) is a **soft memory ceiling** — the GC works harder as you approach it. You usually leave both alone and change them only from data: raise `GOGC` to trade memory for less GC CPU on a GC-bound service; set `GOMEMLIMIT` in containers (below the cgroup limit) to avoid OOM-kills. Tune via env var, not code.
</details>

<details>
<summary><strong>5. Design a job queue: what makes it production-grade (3 things)?</strong></summary>

(1) **Retries with exponential backoff + jitter** so transient failures recover without hammering downstreams or synchronizing into a retry storm. (2) A **dead-letter queue**: after `maxRetries`, move the job aside so one poison job doesn't loop forever or block others. (3) **Idempotent handlers**, because at-least-once execution means a job can run more than once. Plus a **bounded** queue (backpressure) and a sized worker pool. Build all of these from day one or you'll bolt them on after an incident.
</details>

<details>
<summary><strong>6. Why backoff *and* jitter on retries?</strong></summary>

**Backoff** (growing delay between attempts) stops retries from hammering a struggling downstream, giving it time to recover. **Jitter** (randomness added to each delay) de-synchronizes clients: without it, many clients that failed at the same moment retry at the same moment, producing periodic thundering herds that re-trigger the outage. Together they spread retries out in both magnitude and time — the AWS-recommended pattern ("Exponential Backoff and Jitter").
</details>

<details>
<summary><strong>7. Why must job handlers be idempotent?</strong></summary>

Because job execution is at-least-once: a job can be retried after a transient failure, re-run after a worker crash before it acked, or redelivered by a durable queue. If the handler isn't idempotent, those duplicates double-apply the effect (double email, double charge, double shipment). An idempotent handler — dedupe on a job ID, or use naturally idempotent operations / upserts — produces the same result whether the job runs once or several times.
</details>

<details>
<summary><strong>8. What's the optimization workflow, and why "algorithmic before micro"?</strong></summary>

Profile under realistic load → find the hot spot (`top`, `list`) → apply the right fix → re-benchmark to prove the win (`benchstat`) → stop at "good enough." Algorithmic wins (O(n) vs O(n²), batching, caching, eliminating an N+1 query) dwarf micro-optimizations: no amount of `sync.Pool` recovers the cost of a quadratic loop or a per-row DB round-trip. Fix the algorithm first; reach for allocation and micro tweaks only after.
</details>

<details>
<summary><strong>9. How many workers should a pool have?</strong></summary>

Depends on the work. **CPU-bound** work caps useful parallelism around the number of cores (`runtime.NumCPU()`) — more goroutines just add scheduling overhead. **I/O-bound** work (network, DB) can use many more workers since each spends most of its time waiting. Also bound the pool to **protect downstreams**: unbounded concurrency against a database or API can overwhelm it. Tune from measured throughput/latency, and use the bound as a backpressure/failure-isolation control, not just a speed knob.
</details>

<details>
<summary><strong>10. What does "distributed" add to an in-process job queue?</strong></summary>

The logic is the same (queue, workers, retries+backoff, DLQ, idempotency); "distributed" swaps the in-memory channel for **durable, shared storage** — a Redis list or a Kafka topic — and runs workers across multiple machines. That buys durability (jobs survive a process restart), horizontal scaling (add worker nodes), and at-least-once delivery from the broker. It also forces the production concerns: idempotent handlers (duplicates are now guaranteed), consumer-lag monitoring, and backpressure on intake.
</details>

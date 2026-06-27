# Day 29 — Performance Optimization & a Distributed Job Queue

> Mentor note: Day 10 taught you to *measure*. Today you learn what to *do* about it: where Go spends memory, how to cut allocations, and how the GC behaves — plus you build a real **job queue** with workers, retries, and backoff, the kind of thing every backend eventually needs. The performance rule still holds: profile first. But knowing the common fixes means that when the profiler points at allocations, you know exactly which lever to pull.

---

## 1. Where Go performance goes: allocations & the heap

Go is garbage-collected. The GC is excellent (low-latency, concurrent) but not free — every heap allocation is future GC work. **Reducing allocations is the single biggest lever** for most Go services. To reduce them you need to know where they come from.

### Escape analysis — stack vs heap
The compiler puts a value on the **stack** (free, auto-reclaimed) if it can prove the value doesn't outlive the function. If it might escape (returned by pointer, stored in an interface, captured by a closure that outlives the call), it goes on the **heap** (GC-managed). See what escapes:

```bash
go build -gcflags='-m' ./...     # prints "escapes to heap" decisions
```

Common causes of escape: returning a pointer to a local, putting a value in an `interface{}` (boxing), `fmt` with `...interface{}`, growing slices/maps, closures capturing variables.

### Cutting allocations
- **Pre-size** slices/maps (`make([]T, 0, n)`) — Day 2/10.
- **`strings.Builder`** not `+=` in loops — Day 10.
- **Avoid `interface{}` boxing** in hot loops; generics (Day 8) keep types concrete.
- **Reuse buffers** with `sync.Pool`.
- Pass small structs by value (stack, cache-friendly); pass big ones by pointer — but pointers can force heap escape, so measure.

### sync.Pool — reuse instead of reallocate
For frequently allocated, short-lived objects (buffers, encoders) on a hot path:

```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func handle() {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufPool.Put(buf)   // return it for reuse
    // ... use buf
}
```
`sync.Pool` is GC-aware (it's cleared on GC), so it's a cache, not a guarantee. Use it only where a profile shows allocation pressure — it adds complexity and is wrong for long-lived objects.

**Senior take:** `sync.Pool` is a precision tool, not a default. Reach for it when pprof's alloc profile shows a hot object being allocated millions of times. Misused (pooling long-lived or rarely-allocated objects) it adds bugs (forgetting to Reset → data leaks between requests) for no gain. Measure, then pool.

## 2. GC tuning
- **`GOGC`** (default 100) controls the heap-growth target: GC runs when the heap grows GOGC% since the last collection. Higher `GOGC` (e.g. 200) = fewer GCs, more memory; lower = more GCs, less memory. Tune via the `GOGC` env var, not code.
- **`GOMEMLIMIT`** (Go 1.19+) sets a soft memory ceiling — invaluable in containers to avoid OOM-kills; the GC works harder as you approach it.
- Usually you don't touch these. When you do, it's because a profile/observability told you GC is a measurable cost. Throwing memory (higher GOGC) at a CPU-bound-on-GC service is a legitimate, common tuning.

## 3. The optimization workflow (unchanged from Day 10, reinforced)
1. **Profile** (`pprof` CPU + alloc) under realistic load.
2. **Find the hot spot** (`top`, `list`).
3. **Apply the right fix** (algorithmic > allocation > micro).
4. **Re-benchmark** to prove the win (benchstat).
5. Stop when it's good enough — diminishing returns are real.

Algorithmic wins (O(n) vs O(n²), batching, caching — Day 28) dwarf micro-optimizations. Don't `sync.Pool` your way around an N+1 query.

## 4. Project: a distributed job queue

The pattern behind background processing (emails, image resizing, report generation): a queue of jobs, a pool of workers, **retries with backoff** for transient failures, and a **dead-letter** path for jobs that keep failing. (In-process here; "distributed" means backing the queue with Redis/Kafka and running workers across machines — same logic, durable storage.)

Core pieces:
- **Job**: an ID, a payload, an attempt count, max retries.
- **Queue**: a buffered channel (in-process) or Redis list / Kafka topic (distributed).
- **Workers**: a bounded pool (Day 14) pulling jobs.
- **Retry with exponential backoff + jitter**: on failure, requeue with a growing delay (`base * 2^attempt` + random jitter) up to `maxRetries`, then dead-letter. Jitter prevents synchronized retry storms (the Day 1 retry lesson, distributed).
- **Idempotency** (Day 27/28): a job may run more than once — make handlers idempotent.

**Senior take:** Three things separate a toy queue from a real one: **retries with backoff+jitter** (so transient failures recover without hammering), a **dead-letter queue** (so one poison job doesn't loop forever or block others), and **idempotent handlers** (because at-least-once means a job can run twice). Build all three from day one or you'll add them after an incident.

## Common mistakes
1. Optimizing without profiling (Day 10's sin, repeated).
2. `sync.Pool` for long-lived or rarely-allocated objects, or forgetting to `Reset` pooled buffers.
3. Retries with no backoff/jitter → retry storms that amplify an outage.
4. No dead-letter → poison jobs loop forever, blocking the queue.
5. Non-idempotent job handlers under at-least-once execution.
6. Unbounded job intake → memory blowup (no backpressure).
7. Tuning `GOGC`/`GOMEMLIMIT` blindly instead of from data.

## Performance
- The biggest wins are usually fewer allocations and better algorithms, not lower-level tricks.
- Worker count: CPU-bound ≈ cores; I/O-bound higher. Bound it to protect downstreams.
- Backoff base/cap and jitter shape your recovery behavior under load — tune with the failure mode in mind.

---

## Expert Thinking Mode — "make it faster / handle background work"

- **Beginner:** "Spawn a goroutine per job; rewrite the slow-looking loop."
- **Senior:** "Profile → cut allocations / fix the algorithm → re-benchmark. Bounded worker pool, retries with backoff+jitter, DLQ, idempotent handlers."
- **Staff:** "GC cost vs memory trade (GOGC/GOMEMLIMIT) from observability; p99 under load not average; durable distributed queue (Redis/Kafka) with at-least-once + idempotency; backpressure and lag monitoring."
- **Architect:** "Throughput, latency SLOs, and cost are designed together; the queue is part of the system's failure-isolation and scaling strategy; capacity planning for retries and spikes."

---

## Real-world use

- **sync.Pool** is used in `encoding/json`, HTTP/2 framing, and high-throughput Go services to cut GC pressure.
- **GOMEMLIMIT** is now standard in containerized Go to prevent OOM-kills.
- **Job queues** (Sidekiq-style) back every async workload; Go shops build them on Redis (asynq, machinery) or Kafka.
- **Backoff+jitter** is the AWS-blessed standard for retries ("Exponential Backoff and Jitter").

---

## Interview Questions

1. Stack vs heap in Go — what decides, and how do you see escape decisions?
2. Name four ways to reduce allocations.
3. When is `sync.Pool` appropriate and when is it a mistake?
4. What do `GOGC` and `GOMEMLIMIT` control? When would you change them?
5. Design a job queue: what makes it production-grade (3 things)?
6. Why backoff *and* jitter on retries?
7. Why must job handlers be idempotent?

---

## Project (in `../exercises/` and `../solutions/`)

Build the **job queue**: a bounded worker pool processing jobs, with **exponential backoff + jitter retries** up to `maxRetries`, and a **dead-letter** collection for jobs that exhaust retries. A flaky job handler (fails the first couple of attempts) proves the retry path; a permanently-failing job proves the dead-letter path. Run it and report how many jobs succeeded vs dead-lettered, and the attempt counts. Reference in `../solutions/`.

---

## Day 29 companion files

Self-study companions for this day (in `../`):

- [`debugging/`](../debugging/) — the hot-path-allocates-a-buffer bug (fixed with `sync.Pool`), including a `-benchmem` benchmark, with `bugged/` and `fixed/`.
- [`PITFALLS.md`](../PITFALLS.md) — performance & job-queue gotchas as Trap → Why → Fix.
- [`INTERVIEW.md`](../INTERVIEW.md) — interview questions with model answers.
- [`NOTES.md`](../NOTES.md) — quick reference + key terms.
- [`RESOURCES.md`](../RESOURCES.md) — curated links (GC guide, pprof, sync.Pool, backoff+jitter).

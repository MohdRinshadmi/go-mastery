# Day 29 — Quick Reference (performance & job queues)

## Where Go performance goes: allocations
GC is excellent but not free — every heap allocation is future GC work. **Reducing allocations is the biggest lever.**

### Escape analysis (stack vs heap)
- Stack = free, auto-reclaimed; heap = GC-managed.
- Value escapes if it outlives the function: returned by pointer, stored in `interface{}`, captured by an outliving closure, unbounded size.
- See it: `go build -gcflags='-m' ./...`

### Cutting allocations
- Pre-size: `make([]T, 0, n)`.
- `strings.Builder` over `+=` in loops.
- Avoid `interface{}` boxing in hot loops (generics keep types concrete).
- Reuse buffers with `sync.Pool`.
- Small structs by value (stack-friendly); big ones by pointer (but pointers can escape — measure).

### sync.Pool
```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()              // CRUCIAL — clear old data
defer bufPool.Put(buf)
```
> Precision tool, not a default. Only for short-lived objects allocated millions of times (profile says so). GC-aware (a cache, not a guarantee). Always `Reset`; copy out anything the caller keeps.

## GC tuning (from data, not vibes)
- **`GOGC`** (default 100): GC when heap grows GOGC% since last collection. Higher = fewer GCs, more memory.
- **`GOMEMLIMIT`** (1.19+): soft memory ceiling; GC works harder near it. Set in containers below the cgroup limit to avoid OOM-kills.
- Tune via env var, not code; only when observability shows GC is a real cost.

## Optimization workflow
1. Profile (`pprof` CPU + alloc) under realistic load.
2. Find the hot spot (`top`, `list`).
3. Apply the fix — **algorithmic > allocation > micro**.
4. Re-benchmark (`benchstat`) to prove the win.
5. Stop at good enough. (Don't `sync.Pool` your way around an N+1 query.)

## Job queue (production-grade = 3 things)
- **Job**: ID, payload, attempt count, maxRetries.
- **Queue**: buffered channel (in-proc) or Redis list / Kafka topic (distributed) — **bounded** for backpressure.
- **Workers**: bounded pool. CPU-bound ≈ cores; I/O-bound higher; bound to protect downstreams.
- **Retries**: exponential backoff (`base * 2^attempt`) **+ jitter**, up to maxRetries → then **dead-letter**.
- **Idempotent handlers** (at-least-once → a job can run twice).

## Key terms
**Escape analysis** · **stack vs heap** · **allocs/op** · **`sync.Pool`** (+ `Reset`) · **`GOGC`** · **`GOMEMLIMIT`** · **backoff + jitter** · **dead-letter queue** · **backpressure / bounded queue** · **idempotent handler** · **benchstat**.

# Day 28 — Caching Patterns & Distributed Systems Fundamentals

> Mentor note: Caching looks trivial — "store the answer so you don't recompute it" — and it is, until you have many servers, concurrent writes, and a cache that can lie to your users. Today is about caching done *correctly* (the patterns, and the failure modes that cause real outages: stampedes, stale data, thundering herds) plus the distributed-systems vocabulary you'll be quizzed on in every senior interview: CAP, consistency, idempotency. This is where "it works on one box" meets "it works on a hundred boxes."

---

## 1. Why cache
A read from Redis (in-memory, ~0.2ms) vs Postgres (disk + query, ~5–50ms) is 1–2 orders of magnitude faster, and it offloads your database so it survives traffic spikes. Cache the **read-heavy, expensive, slowly-changing** things. Don't cache cheap or rapidly-changing data — the bookkeeping costs more than it saves.

## 2. Caching patterns

### Cache-aside (lazy loading) — the default
The application manages the cache:
1. Read cache. **Hit** → return.
2. **Miss** → read DB, write to cache with a **TTL**, return.
3. On update → **invalidate** (delete) the key.

```go
func GetProduct(ctx context.Context, id string) (Product, error) {
    if p, ok := cache.Get(ctx, id); ok {
        return p, nil               // hit
    }
    p, err := db.GetProduct(ctx, id) // miss -> source of truth
    if err != nil {
        return Product{}, err
    }
    cache.Set(ctx, id, p, 5*time.Minute) // populate with TTL
    return p, nil
}
```
Pros: only caches what's actually read; resilient (cache down → still works via DB). Cons: first read is a miss; risk of stale data between update and invalidation.

### Write-through
Write to cache and DB together on every write. Cache always fresh; writes slower. Good for read-heavy data that must be current.

### Write-behind (write-back)
Write to cache, async-flush to DB later. Fast writes, risk of data loss if cache dies before flush. Use rarely and carefully.

**Senior take:** Default to **cache-aside with a TTL**. The TTL is your safety net: even if an invalidation is missed (bug, race, another service wrote directly), stale data self-heals within the TTL. A cache entry with no TTL is a bug — it's both a memory leak and an indefinitely-stale value.

## 3. The failure modes (this is the real lesson)

### Cache stampede / thundering herd
A hot key expires. 10,000 concurrent requests all miss simultaneously and all hit the DB at once → DB falls over. Defenses:
- **Single-flight**: only one goroutine recomputes; the rest wait for its result (`golang.org/x/sync/singleflight`). Within one process this collapses N concurrent misses into 1 DB call.
- **Locking / lease**: first miss takes a short lock (e.g. a Redis `SETNX`), recomputes, others briefly serve stale or wait.
- **Staggered/jittered TTLs**: don't expire many keys at the same instant; add random jitter so expiries spread out.
- **Refresh-ahead**: proactively refresh hot keys before they expire.

### Stale data / invalidation
"There are only two hard things in CS: cache invalidation and naming things." When the DB changes, the cache must be invalidated — but across services and replicas that's racy. TTLs bound the staleness; for stronger freshness use write-through or event-driven invalidation (publish an "invalidate key X" event — Day 27).

### Other traps
- **Cache penetration**: requests for keys that don't exist (often malicious) always miss and hit the DB. Cache the "not found" result (with a short TTL) too.
- **Hot key**: one key gets disproportionate traffic, overloading one Redis shard. Replicate or local-cache it.

## 4. Distributed-systems fundamentals (interview core)

### CAP theorem
In a network partition (P — unavoidable in distributed systems), you must choose between:
- **C**onsistency — every read sees the latest write (or an error).
- **A**vailability — every request gets a (possibly stale) response.

You can't have both *during a partition*. Real systems pick a lean: Postgres (CP-ish), Cassandra/Dynamo (AP, eventual consistency), etc. **There is no "CA" system** in the real world because partitions happen. Caches are inherently an availability/consistency trade — a cache is "eventually consistent" with the DB.

### Consistency models (a spectrum)
- **Strong** — reads always reflect the latest write. Simple to reason about, costly to scale.
- **Eventual** — replicas converge "eventually"; reads may be stale briefly. Scales well (most caches, AP datastores).
- **Read-your-writes / causal** — useful middle grounds (you at least see your own updates).

### Idempotency (again, because it's everywhere)
An operation safe to apply multiple times with the same effect. Critical in distributed systems because retries, redelivery (Day 27), and partitions cause duplicates. **Idempotency keys** (a client-supplied unique ID per logical operation) let a server dedupe retried requests — exactly how Stripe's API prevents double charges.

### Other terms you'll be asked
- **Replication** (copies for availability/reads) vs **sharding/partitioning** (split data for scale).
- **Quorum** (R + W > N) for tunable consistency in replicated stores.
- **Two-phase / saga** for distributed transactions (sagas + compensation are preferred over 2PC at scale).

**Senior take:** Most "distributed systems" interview failures are conflating consistency with availability, or claiming exactly-once delivery exists. Anchor on: partitions are inevitable → you trade C for A → therefore design for eventual consistency and idempotent operations. That single chain answers half the questions.

## Common mistakes
1. Cache entries with no TTL → stale forever + memory leak.
2. No stampede protection on hot keys → DB meltdown on expiry.
3. Treating a local in-process cache as cluster-wide (each replica has its own).
4. Caching write paths or rapidly-changing data (low hit rate, high staleness).
5. Not caching "not found" → cache penetration.
6. Claiming strong consistency from an eventually-consistent cache.
7. Non-idempotent operations behind retries.

## Performance
- Measure **hit rate** — a cache below ~80–90% hit rate on hot data may not be worth it; profile.
- `singleflight` collapses in-process stampedes for near-free; distributed stampedes need Redis-level coordination.
- Serialization cost (JSON/gob) is part of cache latency; for tiny values it can dominate — measure.

---

## Expert Thinking Mode — "make reads faster"

- **Beginner:** "Add a global map cache." (Stale, per-pod, no eviction.)
- **Senior:** "Cache-aside + TTL, invalidate on write, singleflight for stampedes, cache negatives. Know the hit rate."
- **Staff:** "Consistency requirement drives the pattern (cache-aside vs write-through vs event-driven invalidation). Hot-key and stampede mitigation. Local + distributed cache tiers. CAP trade made explicit."
- **Architect:** "Caching is a consistency-vs-latency-vs-cost system decision tied to the data's freshness SLA. Invalidation topology, cache-coherence across services, and failure behavior (cache down ⇒ graceful degradation) are designed up front."

---

## Real-world use

- **Redis** in front of Postgres/MySQL is the most common read-scaling pattern in web backends.
- **Stripe idempotency keys** prevent duplicate charges on client retries — textbook distributed idempotency.
- **singleflight** is used inside Go services (and Go's own DNS resolver) to dedupe concurrent identical work.
- **CAP/eventual consistency** underpins Dynamo/Cassandra and every CDN/cache layer (Cloudflare).

---

## Interview Questions

1. Describe cache-aside. Why is a TTL essential even with invalidation on write?
2. What is a cache stampede and how do you prevent it (name 3 techniques)?
3. Write-through vs cache-aside — trade-offs?
4. State the CAP theorem. Why is there no CA system in practice?
5. Strong vs eventual consistency — give an example system of each.
6. What is an idempotency key and what problem does it solve?
7. What is cache penetration and how do you defend against it?

---

## Your tasks

`../exercises/` has a slow "database" (artificial latency) and asks you to: (1) implement a cache-aside `Cache` with per-entry TTL, (2) add **singleflight** so concurrent misses for the same key cause exactly ONE DB call (the demo fires 50 concurrent gets and counts DB calls — prove it's 1), and (3) cache negative lookups. The runnable demo uses an in-memory cache; the real go-redis version is in `solutions/redis_reference.go` (build-ignored). Reference in `../solutions/`.

---

## Day 28 companion files

Self-study companions for this day (in `../`):

- [`debugging/`](../debugging/) — the cache stampede bug (cache-aside without single-flight) with `bugged/` and `fixed/`.
- [`PITFALLS.md`](../PITFALLS.md) — caching/distributed gotchas as Trap → Why → Fix.
- [`INTERVIEW.md`](../INTERVIEW.md) — interview questions with model answers.
- [`NOTES.md`](../NOTES.md) — quick reference + key terms.
- [`RESOURCES.md`](../RESOURCES.md) — curated links (Redis patterns, singleflight, CAP).

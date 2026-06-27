# Day 28 — Quick Reference (caching & distributed systems)

## Why cache
Redis (~0.2ms) vs Postgres (~5–50ms) = 1–2 orders of magnitude faster + offloads the DB.
Cache **read-heavy, expensive, slowly-changing** data. Don't cache cheap or fast-changing data.

## Patterns
| Pattern | How | Trade-off |
|---|---|---|
| **Cache-aside** (default) | app reads cache → miss → DB → populate w/ TTL → invalidate on write | first read misses; staleness window |
| Write-through | write cache + DB together | always fresh; slower writes |
| Write-behind | write cache, async flush to DB | fast writes; data-loss risk |

```go
func Get(id string) (Product, error) {
    if p, ok := cache.Get(id); ok { return p, nil } // hit
    p, err := db.Get(id); if err != nil { return Product{}, err }
    cache.Set(id, p, 5*time.Minute) // populate w/ TTL
    return p, nil
}
```
> **A cache entry with no TTL is a bug** — memory leak + indefinitely stale.

## Failure modes
- **Stampede / thundering herd** — hot key expires, N misses hammer the DB. Fix: single-flight (in-proc), Redis lock/lease (cross-proc), **jittered TTLs**, refresh-ahead.
- **Stale data / invalidation** — racy across replicas; TTL bounds it; write-through or event-driven invalidation for stronger freshness.
- **Cache penetration** — lookups for non-existent keys always miss → cache the "not found" (short TTL) / Bloom filter.
- **Hot key** — one key overloads one shard → replicate or local-cache it.

## Distributed-systems fundamentals
- **CAP** — under a partition (P, inevitable) pick **C** or **A**. No CA in practice. Caches trade C for A (eventually consistent).
- **Consistency models** — strong (latest write) · eventual (converges, may be stale) · read-your-writes / causal (middle).
- **Idempotency** — operation safe to apply many times; **idempotency keys** dedupe retries (Stripe).
- **Replication** (copies, availability/reads) vs **sharding** (split, scale).
- **Quorum** R+W>N for tunable consistency; **saga + compensation** over 2PC for distributed txns.

## Anchor chain (half the interview)
Partitions are inevitable → trade C for A → design for **eventual consistency + idempotent operations**.

## Key terms
**Cache-aside / lazy loading** · **TTL** · **Stampede / thundering herd** · **Single-flight** · **Negative caching / cache penetration** · **Hot key** · **CAP** · **Strong vs eventual consistency** · **Idempotency key** · **Replication vs sharding** · **Quorum** · **Saga**.

> Runnable exercises use an in-memory cache so they work offline; the go-redis version is in `solutions/redis_reference.go` (build-ignored). Real single-flight: `golang.org/x/sync/singleflight`.

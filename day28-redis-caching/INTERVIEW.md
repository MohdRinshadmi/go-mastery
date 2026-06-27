# Day 28 — Interview Q&A (caching & distributed systems)

<details>
<summary><strong>1. Describe cache-aside. Why is a TTL essential even with invalidation on write?</strong></summary>

Cache-aside (lazy loading): the app checks the cache; on a hit it returns; on a miss it reads the DB, populates the cache with a TTL, and returns; on a write it invalidates (deletes) the key. The TTL is essential because invalidation is never 100% reliable across services and replicas — a bug, a race, or another service writing directly to the DB can leave a stale entry. The TTL bounds staleness: even a missed invalidation self-heals within the TTL. An entry with no TTL is both a memory leak and an indefinitely-stale value.
</details>

<details>
<summary><strong>2. What is a cache stampede and how do you prevent it (name 3 techniques)?</strong></summary>

A stampede (thundering herd) is when a hot key expires and many concurrent requests all miss simultaneously and all hit the DB at once, potentially overwhelming it. Defenses: (1) **single-flight** — only one goroutine recomputes per key, the rest wait for its result (in-process); (2) **locking/lease** — first miss takes a short distributed lock (Redis `SETNX`), others wait or serve stale (cross-process); (3) **jittered TTLs** — add randomness so many keys don't expire at the same instant; plus (4) **refresh-ahead** — proactively refresh hot keys before they expire.
</details>

<details>
<summary><strong>3. Write-through vs cache-aside — trade-offs?</strong></summary>

Cache-aside populates lazily on read and invalidates on write: only caches what's actually read, resilient (cache down → app still works via DB), but the first read is a miss and there's a staleness window between a write and its invalidation. Write-through writes to cache and DB together on every write: the cache is always fresh and reads never serve stale data, but writes are slower and you cache things that may never be read. Default to cache-aside + TTL; use write-through for read-heavy data that must always be current.
</details>

<details>
<summary><strong>4. State the CAP theorem. Why is there no CA system in practice?</strong></summary>

In a network partition (P), a distributed system must choose between Consistency (every read sees the latest write, or errors) and Availability (every request gets a — possibly stale — response). You can't have both *during a partition*. There's no real "CA" system because partitions are inevitable in any real network; the moment one happens you're forced to pick C or A. So "CA" only describes a single non-distributed node. Real systems lean CP (Postgres) or AP (Cassandra/Dynamo) by how they behave under partition.
</details>

<details>
<summary><strong>5. Strong vs eventual consistency — give an example system of each.</strong></summary>

Strong consistency: every read reflects the latest committed write; simple to reason about, costlier to scale — e.g. a single-primary Postgres, or a linearizable store like etcd/ZooKeeper. Eventual consistency: replicas converge "eventually", so reads can be briefly stale; scales well — e.g. Cassandra/DynamoDB, DNS, CDNs, and caches (a cache is eventually consistent with its DB). Middle grounds like read-your-writes and causal consistency give useful guarantees without full strong consistency.
</details>

<details>
<summary><strong>6. What is an idempotency key and what problem does it solve?</strong></summary>

A client-supplied unique ID attached to a logical operation (e.g. an HTTP header `Idempotency-Key`). The server records the result keyed by it, so if the same request is retried (network timeout, client retry, redelivery) the server recognizes the key and returns the original result instead of performing the operation again. It solves the duplicate-execution problem inherent to distributed systems — exactly how Stripe prevents a retried charge from billing the customer twice.
</details>

<details>
<summary><strong>7. What is cache penetration and how do you defend against it?</strong></summary>

Cache penetration is repeated requests for keys that **don't exist** — they always miss the cache and fall through to the DB, which is wasteful and a common DoS vector (attacker requests random non-existent IDs). Defenses: **cache the negative result** ("not found") with a short TTL so repeated misses are absorbed, and/or a **Bloom filter** in front to cheaply reject keys that definitely don't exist before touching the cache or DB.
</details>

<details>
<summary><strong>8. Replication vs sharding (partitioning) — what's the difference?</strong></summary>

Replication keeps **copies** of the same data on multiple nodes — for availability and read scaling (reads can hit any replica; a node can fail without data loss). Sharding/partitioning **splits** the data across nodes by key — for write/storage scale (each shard holds a subset). They're orthogonal and usually combined: shard for capacity, replicate each shard for fault tolerance. Replication raises consistency questions (which replica do you read?); sharding raises routing/rebalancing questions.
</details>

<details>
<summary><strong>9. Why does single-flight not fully solve a distributed stampede?</strong></summary>

`singleflight` collapses concurrent identical work **within a single process**. In a fleet of N replicas, each process independently collapses its own misses — so a hot key expiring can still cause up to N DB loads (one per pod), not one globally. To collapse across processes you need cross-process coordination: a distributed lock/lease (Redis `SETNX` with a TTL) so only one pod recomputes while others wait or serve stale. Single-flight + distributed lock + jittered TTLs together handle both scopes.
</details>

<details>
<summary><strong>10. When should you NOT cache something?</strong></summary>

When the data is cheap to compute/fetch (the cache bookkeeping costs more than it saves), when it changes rapidly (low hit rate + high staleness risk), or when it's write-heavy (you're mostly invalidating, not serving). Caching shines for read-heavy, expensive, slowly-changing data. Always measure the **hit rate**: below ~80–90% on supposedly hot data, the cache may not be earning its complexity.
</details>

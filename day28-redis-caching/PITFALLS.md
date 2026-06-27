# Day 28 — Pitfalls (caching & distributed systems)

Format: **Trap → Why → Fix**.

### 1. Cache entries with no TTL
**Trap:** A key is set once and never expires, serving stale data indefinitely and leaking memory.
**Why:** Without a TTL, a missed invalidation (bug, race, another writer) is never self-corrected.
**Fix:** Always set a TTL. It's the safety net — stale data self-heals within the TTL even if invalidation is missed.

### 2. No stampede protection on hot keys
**Trap:** A hot key expires and thousands of concurrent misses hit the DB at once, melting it.
**Why:** Plain cache-aside has no coordination between concurrent misses for the same key.
**Fix:** Single-flight (in-process), a Redis lock/lease (cross-process), and jittered TTLs so expiries spread out.

### 3. Treating an in-process cache as cluster-wide
**Trap:** Invalidate-on-write only clears the local map; the other 9 replicas still serve the old value.
**Why:** Each pod has its own in-memory cache; there's no shared state.
**Fix:** Use a shared cache (Redis) for cross-replica consistency, or event-driven invalidation (publish "invalidate key X").

### 4. Not caching "not found"
**Trap:** Requests for keys that don't exist always miss and hit the DB (cache penetration; often malicious).
**Why:** A miss isn't cached, so every lookup for a non-existent key reaches the source.
**Fix:** Cache the negative result with a *short* TTL so repeated lookups for missing keys are absorbed.

### 5. Caching a miss/error as a real value
**Trap:** A transient DB error or empty result gets stored as if it were the answer, then served for the whole TTL.
**Why:** Cache-aside that doesn't distinguish "real empty" from "failed to load" poisons the cache.
**Fix:** Only cache successful, authoritative results. Cache negatives deliberately and with a short TTL, never errors.

### 6. Claiming strong consistency from a cache
**Trap:** Code (or an interviewer answer) assumes reads always reflect the latest write through a cache.
**Why:** A cache is eventually consistent with the DB; there's always a staleness window.
**Fix:** Design for eventual consistency; if you truly need fresh reads use write-through or read past the cache.

### 7. Non-idempotent operations behind retries
**Trap:** Retried/duplicated requests (network, redelivery, partitions) double-apply effects.
**Why:** Distributed systems produce duplicates; non-idempotent ops can't tolerate them.
**Fix:** Idempotency keys (client-supplied unique ID per logical op) so the server dedupes retries — the Stripe pattern.

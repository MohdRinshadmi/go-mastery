# Day 28 debugging — the cache stampede that took down the database

**Phase 6 · caching · stampede / thundering herd**

> Stdlib only. The "cache" is a mutex-guarded map and the "DB" is a slow function,
> so the failure mode (concurrent misses overwhelming the source of truth) is real
> and reproducible offline.

## Symptom

A cache-aside `Get` works perfectly in single-threaded tests. In production,
every time a popular product's cache entry is cold — the first request after a
deploy, or the instant a hot key expires — the database gets a burst of identical
queries and CPU spikes. The cache "works", yet the DB keeps getting hammered for
the *same* key.

```bash
cd bugged
go run -race .
```

Expected: 50 concurrent requests for one cold key → **1** DB call.
Actual: ~50 DB calls — every request missed and loaded independently.

## Hint

Cache-aside is correct for *one* caller at a time: check cache → miss → load DB →
populate. Now picture 50 goroutines hitting a cold key at the *same instant*. They
all run the check, they all see "not cached" (nobody's populated it yet), and they
all proceed to the DB. The window between "I missed" and "I populated the cache"
is open to everyone. What collapses many concurrent misses for the same key into a
single load?

## How to reproduce

`go run -race .` in `bugged/`. It fires 50 goroutines at one cold key and counts
how many times the slow DB was actually called.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

Plain cache-aside has no coordination between concurrent misses:

```go
if v, ok := s.c.get(key); ok { return v } // all 50 see "miss"
v := s.db.load(key)                        // all 50 hit the DB
s.c.set(key, v)                            // all 50 populate (redundantly)
```

This is the **cache stampede** (a.k.a. thundering herd). A single hot key
expiring turns into N simultaneous identical DB queries. The cache doesn't help
at the one moment you need it most — under load, on your hottest key. At real
scale (10,000 concurrent requests) this is a self-inflicted DoS on your database.

### The fix

**Single-flight**: ensure only one goroutine loads a given key at a time; the rest
wait for and share its result.

```go
s.mu.Lock()
if cl, ok := s.inFlight[key]; ok { // someone is already loading it
    s.mu.Unlock()
    cl.wg.Wait()                   // wait for their result
    return cl.val
}
cl := &call{}; cl.wg.Add(1); s.inFlight[key] = cl
s.mu.Unlock()
cl.val = s.db.load(key)            // exactly one DB call per key
s.c.set(key, cl.val)
// ... delete from inFlight, cl.wg.Done() to release waiters
```

In real code you use `golang.org/x/sync/singleflight` — same idea, battle-tested:

```go
v, err, _ := group.Do(key, func() (any, error) { return db.Load(key) })
```

Important scope note:

> - `singleflight` collapses stampedes **within one process**. With many replicas,
>   each process still does one load — so a fleet of 100 pods can still issue 100
>   DB calls. For cross-process stampedes you also need a **distributed lock**
>   (Redis `SETNX`/lease) so only one pod recomputes.
> - Complementary defenses: **jittered TTLs** (so many keys don't expire at the same
>   instant), and **refresh-ahead** (proactively refresh hot keys before expiry).

Rules:

> 1. Cache-aside under concurrency needs stampede protection on hot keys. Correct
>    single-threaded ≠ safe under load.
> 2. `singleflight` is in-process; distributed stampedes need a Redis-level lock.
> 3. Add TTL jitter so expiries spread out instead of synchronizing.

`fixed/` collapses 50 concurrent misses into 1 DB call; the bugged version makes ~50.

</details>

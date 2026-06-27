# Day 13 debugging — the metrics store that races (and crashes)

**Phase 3 · Concurrency · data race on a shared map**

## Symptom

An in-memory metrics store: many goroutines call `s.Inc(key)` concurrently to
count events. The totals should sum to the number of `Inc` calls. They don't —
and sometimes the program **crashes outright**. Run it:

```bash
cd bugged
go run .
```

You'll see one of:

```
recorded 10000 hits, store totals 9871
BUG: 129 updates lost to the data race
```

or a hard crash:

```
fatal error: concurrent map writes
```

Either way the store is unreliable. The lost-update version is the scary one: it
looks like it worked, just with subtly wrong numbers.

## Hint

Let the tool name it:

```bash
go run -race .
```

Read the `WARNING: DATA RACE` block — it points at `m[key]++` and the unguarded
read in `Total()`. Two things are wrong: (1) `m[key]++` is load-add-store, not
atomic, so concurrent increments collapse; (2) writing a Go `map` from multiple
goroutines is *itself* illegal and the runtime may `fatal error: concurrent map
writes`.

## How to reproduce

`go run -race .` in `bugged/` — prints `WARNING: DATA RACE` every run (often
several). Without `-race`, you'll see wrong totals and intermittent
`fatal error: concurrent map writes` crashes.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The `map[string]int` is read and written by many goroutines with **no
synchronization**. Two distinct bugs in one:

1. **Data race / lost updates.** `m[key]++` is read-modify-write. Two goroutines
   read `5`, both compute `6`, both store `6` — one increment vanishes. The total
   drifts below the true count, non-deterministically.
2. **Concurrent map writes.** Go maps are *not* safe for concurrent writes by
   design. The runtime actively detects this and calls `fatal error: concurrent
   map writes` — a crash, not a corruption. (This detection is always on, even
   without `-race`.)

A data race is undefined behavior, so you can't reason about the result at all —
the race detector exists precisely because these bugs are invisible to the eye.

### The fix

Guard the map with a `sync.RWMutex` that lives **next to the data** and is
encapsulated inside the methods:

```go
type Store struct {
    mu sync.RWMutex
    m  map[string]int
}
func (s *Store) Inc(key string) { s.mu.Lock();  defer s.mu.Unlock();  s.m[key]++ }
func (s *Store) Total() int     { s.mu.RLock(); defer s.mu.RUnlock(); /* sum */ }
```

- Writers take `Lock` (exclusive); the read-only `Total` takes `RLock`, allowing
  many concurrent readers — a good fit for read-mostly state.
- `defer` the unlock so the lock releases even if the body panics.
- The mutex is **unexported and owned by the struct** — callers can't forget to
  lock, because locking is internal to the methods.

`fixed/` also threads a `context` with `defer cancel()` so the workload is
cancellable — the Day 13 discipline for any long-running concurrent work.

Alternatives: `sync.Map` (benchmark first — a plain map + `RWMutex` often wins),
or `atomic.Int64` per counter if the key set is fixed.

### The rules

> 1. A `map` written by more than one goroutine **must** be synchronized — there
>    are no exceptions, and the runtime will crash you if you don't.
> 2. Keep the mutex **unexported, next to its data, and locked only inside the
>    methods**. Lock for the shortest span possible; always `defer Unlock`.
> 3. Run any package with goroutines under `-race` in CI. A clean `-race` run is
>    the only proof your synchronization is correct.

Verify: `go run -race .` in `fixed/` prints exact totals with no warning.

</details>

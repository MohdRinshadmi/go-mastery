# Day 13 — Context, Mutexes, Race Detector: Quick Reference

## context

```go
ctx, cancel := context.WithCancel(parent)
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
ctx, cancel := context.WithDeadline(parent, t)
defer cancel()                  // ALWAYS — releases timer/goroutine

ctx = context.WithValue(parent, key, val) // request-scoped data only
```

Consume cancellation:
```go
select {
case res := <-work:  return res, nil
case <-ctx.Done():   return zero, ctx.Err() // Canceled | DeadlineExceeded
}
```

Rules: `ctx` first param, named `ctx` · never store in a struct · always `defer
cancel()` · never pass `nil` (use `Background()` / `TODO()`) · `Value` ≠ arg passing.

## Mutex

```go
type Counter struct {
    mu sync.Mutex
    n  int
}
func (c *Counter) Inc()   { c.mu.Lock(); defer c.mu.Unlock(); c.n++ }
func (c *Counter) Value() int { c.mu.Lock(); defer c.mu.Unlock(); return c.n }
```

- Mutex **unexported, next to its data**, locked **inside** the methods.
- Always `defer Unlock` · lock the shortest span · do slow I/O outside the lock.
- Never copy a mutex (use pointer receivers; `go vet` copylocks catches copies).

## RWMutex

```go
mu.RLock(); /* read */ mu.RUnlock()  // many concurrent readers
mu.Lock();  /* write */ mu.Unlock()  // one exclusive writer
```
Wins only for read-mostly state with a non-trivial critical section; measure.

## atomic (single value, lock-free)

```go
var n atomic.Int64
n.Add(1)
n.Load()
n.CompareAndSwap(old, new)
```
Best for a lone counter/flag. Multiple variables that must change together → mutex.

## Maps and concurrency

- Plain `map` + concurrent writes → data race **and** `fatal error: concurrent
  map writes` (runtime crash, always on).
- Fix: `map` + `RWMutex`, or `sync.Map` (benchmark; map+RWMutex often wins).

## Race detector

```bash
go run -race .
go test -race ./...
```
- Reports the two racing stack traces. ~5–10× overhead → test/CI only, not prod.
- Any package with goroutines → `-race` in CI, no exceptions.

## Channel vs mutex vs atomic

| Need | Use |
|---|---|
| transfer/coordinate data between goroutines | channel |
| protect shared state (multi-field invariant) | mutex |
| read-mostly shared state, non-trivial reads | RWMutex |
| one counter/flag/pointer | atomic |

---

## Key terms

- **context.Context** — cancellation + deadline propagated down a call tree.
- **WithCancel/WithTimeout/WithDeadline** — derive a cancellable child context.
- **cancel()** — releases a context's resources; always `defer` it.
- **ctx.Done()** — channel closed on cancel/timeout; `select` on it.
- **ctx.Err()** — `context.Canceled` or `context.DeadlineExceeded`.
- **data race** — concurrent unsynchronized access, ≥1 write; undefined behavior.
- **sync.Mutex / RWMutex** — exclusive / shared-read locks.
- **atomic** — lock-free indivisible ops on a single value.
- **copylocks** — `go vet` check for copying a mutex/WaitGroup by value.
- **concurrent map writes** — runtime fatal error from unsynchronized map writes.

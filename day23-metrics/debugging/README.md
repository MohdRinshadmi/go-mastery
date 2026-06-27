# Day 23 Debugging — The metrics counter that races (and lies)

A middleware records `http_requests_total{path}` on every request by
incrementing a counter map. It works in a single-threaded test. Under real
concurrent load it either **panics** (`fatal error: concurrent map writes`) or,
worse, silently **undercounts** — your traffic dashboard reads lower than
reality, and you under-provision because the numbers look fine.

We simulate a Prometheus `CounterVec` with a plain map (**stdlib only**, no
`client_golang`). The bug is the missing synchronization.

## Symptom

```
$ cd bugged && go run -race .
==================
WARNING: DATA RACE
Write at 0x... by goroutine 9:
  main.(*counterVec).Inc()
...
http_requests_total{path="/orders"} = 47213 (want 50000)
=> metric UNDERCOUNTED due to the data race; dashboards are wrong
```

(Without `-race` it may instead crash with `fatal error: concurrent map
writes`, or print an undercounted total.)

## Reproduce

```bash
cd bugged
go run -race .     # data race reported; count < want
```

## Hint

<details>
<summary>Hint</summary>

Fifty goroutines call `Inc()` on the same map at the same time. A map in Go is
not safe for concurrent writes, and `x++` on a shared `int64` is a
read-modify-write that loses updates under contention. What guards shared
mutable state across goroutines?

</details>

## Solution & why

<details>
<summary>Solution & why</summary>

`Inc()` did `c.values[label]++` with no lock. Two failures stack up:

1. **Concurrent map access** — Go maps are not safe for concurrent writes; the
   runtime may panic with `concurrent map writes`.
2. **Lost updates** — even ignoring the map, `count++` is load → add → store.
   Two goroutines can read the same old value and both store `old+1`, so one
   increment vanishes. That's the undercount.

The race detector (`-race`) flags the unsynchronized access deterministically,
which is exactly why CI should run `go test -race`.

**Fix:** guard the map with a `sync.Mutex` (or use the real prometheus client,
which uses atomics internally):

```go
func (c *counterVec) Inc(label string) {
    c.mu.Lock()
    c.values[label]++
    c.mu.Unlock()
}
```

Now the count is exact and the race is gone.

**Bonus (the lesson's #1 sin — cardinality):** the `fixed/` version also
normalizes the path to a bounded **route template** (`/orders/123` →
`/orders/{id}`) before recording. Putting raw IDs in a label spawns one time
series per ID and is the classic way to OOM Prometheus. Bounded labels +
thread-safe increment = a metric you can trust.

</details>

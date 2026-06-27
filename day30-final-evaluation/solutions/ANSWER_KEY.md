# Final Exam — Answer Key

## Section A — Theory

1. `string` → `""`, `slice` → `nil`, `map` → `nil`. Reading a nil map is fine; **writing a nil map panics**. nil slices are safe to `append`/`range`.
2. `b` shares `a`'s backing array; `b`'s cap still spans into `a`, so `append` writes into `a`'s index 2 instead of allocating. Prevent with a 3-index slice `a[0:2:2]` (caps capacity) or `copy` into a fresh slice.
3. Failure is explicit and local — no invisible control flow; you see every error path in the signature/next line. Cost: verbosity (`if err != nil` everywhere) and discipline to add context.
4. `%w` wraps (keeps the cause inspectable via `errors.Is/As`); `%v` flattens to a string (hides internal error types — use at an API boundary where you don't want callers depending on the cause).
5. Composition avoids rigid type hierarchies and the fragile base-class problem; small implicit interfaces let any type satisfy a contract without declaring it, so you compose behavior and depend on narrow interfaces.
6. Right: container/algorithm code that's identical across types (Map/Filter/sets/stacks). Hurts: when an interface is clearer, when constraints get complex, or when it adds type-parameter noise for one or two concrete types.
7. Two goroutines access the same memory concurrently, ≥1 writing, with no synchronization → undefined behavior. Unreliable to read-find because it's nondeterministic and usually works in testing; use `go test -race`.
8. Channel: transfer ownership/coordinate/pipeline/signal. Mutex: protect a small piece of shared state with simple read/write. Prefer channels for coordination, mutex when it's simpler/faster.
9. The **sender** closes, exactly once. Misuse: closing from the receiver / closing twice (panic), and sending on a closed channel (panic).
10. Cancellation + deadline (and request-scoped values). Rules: first param named `ctx`; never store in a struct; always `cancel()`; don't pass nil (use `Background`/`TODO`); `Value` only for request-scoped data.
11. Liveness = "is the process alive" → failure **restarts** the pod. Readiness = "can I serve now" (checks deps) → failure **removes from rotation**, no restart. Danger: a DB check in liveness causes restart storms on a DB blip.
12. Under a network **P**artition you must choose **C** (consistency, may error) or **A** (availability, may be stale). No CA in practice because partitions are inevitable, so you can't refuse to tolerate them.
13. **Idempotency** — because a crash between processing and offset-commit causes redelivery (duplicates), the consumer must produce the same effect if it runs twice.
14. Read cache → hit returns; miss reads DB, populates cache with TTL; invalidate on write. TTL is the safety net: missed/raced invalidations self-heal within the TTL (and it bounds memory).
15. Sync gRPC for queries needing an immediate, low-latency, strongly-typed answer (auth check, get product). Async events for fan-out / work that can happen later (OrderPlaced → payment/inventory/email), to decouple and tolerate partial failure.

## Section B — Coding (sketches; see `final_test.go` for runnable B1/B2/B3)

**B1.**
```go
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
    out := make(map[K][]T)
    for _, it := range items {
        k := key(it)
        out[k] = append(out[k], it)
    }
    return out
}
```

**B2.** Map + insertion-order slice (or container/list), mutex around both; on `Set`
beyond capacity, evict `order[0]` and delete from the map. (Runnable version in `final.go`.)

**B3.** Worker pool over a jobs channel of URLs; each worker does
`ctx, cancel := context.WithTimeout(parent, d); err := check(ctx, url); cancel()`;
collect into a mutex-guarded map; bound workers to N; close results after a WaitGroup.

## Section C — Debugging

**C1.** Only the first send is received; the other goroutines block forever on `ch <- ...`
(unbuffered, no more receivers) → leaked goroutines. Fixes: buffer the channel
(`make(chan string, len(urls))`) so all sends complete, **or** use a `context` you
cancel after the first result and `select` on `ctx.Done()` in each goroutine, **or** drain.

**C2.** `counter++` is read-modify-write, not atomic → data race, lost updates.
Fix A (mutex): guard `counter++` with `sync.Mutex`. Fix B (atomic):
`var counter atomic.Int64; counter.Add(1); ... counter.Load()`. Verify with `-race`.

## Section D — Code Review (issues to raise)

1. **SQL injection** — `"... WHERE id = " + id` concatenates user input. Use a parameterized query `WHERE id = $1`.
2. **Ignored `Scan` error** — no error handling; a missing row (`sql.ErrNoRows`) is silently treated as empty → should be 404.
3. **No `QueryRowContext(r.Context())`** — no cancellation/timeout; a slow query hangs the request.
4. **Wrong/implicit status handling** — always returns 200 even on not-found/error; should map errors to 404/500. (Also set `Content-Type: application/json`.)
5. **SQL in the HTTP handler** — no repository/service layer; untestable, unswappable, mixes concerns.
6. **No input validation** — `id` unchecked (empty/format).
7. **Encode error ignored** — `json.NewEncoder(...).Encode(...)` error dropped.
8. (Nit) leaking DB schema/`name` even if user not found; inconsistent error shape; no logging with context.

---

## Self-assessment guide (fill this in)

- **Strongest phases:** ____
- **Weakest 3 areas:** ____  → these are your next 2-week focus.
- **Current band (from rubric):** ____
- **Target role:** Go Backend Engineer / Platform / Distributed Systems
- **Next project to ship:** the Day 20 e-commerce API with real Postgres+Redis+JWT, tests, Docker, CI — deployed.

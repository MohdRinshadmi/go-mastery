# Day 13 — Context, Mutexes, Race Detector: Interview Q&A

Model answers in `<details>`.

---

**1. What problem does `context.Context` solve? Give two concrete uses.**

<details>
<summary>Answer</summary>

It carries a **cancellation signal and deadline** down a call tree, so work that's
no longer needed can stop. Two uses: (1) a server cancels all downstream DB/HTTP
calls when the client disconnects (`r.Context()` fires on disconnect); (2) a
per-request timeout (`context.WithTimeout`) bounds total request time and unblocks
any call selecting on `ctx.Done()`. It also carries request-scoped values (request
ID, auth user) — but that's a secondary, easily-abused feature.
</details>

---

**2. What are the rules for using context?**

<details>
<summary>Answer</summary>

(1) `ctx` is the **first parameter**, named `ctx`. (2) **Don't store it in a
struct** — pass it per call. (3) **Always `defer cancel()`** for any `WithCancel/
Timeout/Deadline` to release its timer/goroutine. (4) **Never pass `nil`** — use
`context.Background()` at the top or `context.TODO()` as a placeholder. (5)
`ctx.Value` is for **request-scoped data only**, not for passing optional function
arguments.
</details>

---

**3. What is a data race? Why can't you reliably find it by reading code?**

<details>
<summary>Answer</summary>

A data race is two goroutines accessing the same memory concurrently, at least one
writing, with no synchronization establishing an order between them. The result is
**undefined** — torn reads, lost updates, corruption, crashes. You can't find it by
reading because it's nondeterministic: the interleaving that triggers it may occur
only under specific timing/load, so the code "works" in dev and fails in prod. The
race detector instruments memory accesses and reports the exact two racing stacks.
</details>

---

**4. How does `go test -race` work, and why isn't it a production build flag?**

<details>
<summary>Answer</summary>

It compiles with instrumentation that records every memory access and the
happens-before relationships from synchronization (channels, mutexes, atomics).
When it sees two unsynchronized accesses to the same address with at least one
write, it reports both stack traces. It adds ~5–10× CPU and significant memory
overhead, so it's a **testing/CI** tool, not a production build. Rule of thumb:
every package with goroutines runs under `-race` in CI.
</details>

---

**5. When do you choose a channel vs a mutex?**

<details>
<summary>Answer</summary>

**Channel** when you're transferring ownership of data or coordinating goroutines
— pipelines, signaling, handing work off. **Mutex** when you have a small piece of
shared state (a counter, a cache, a map) that multiple goroutines read/write in
place. The proverb "share memory by communicating" makes channels the default for
coordination, but don't force a channel where a mutex-guarded field is simpler and
faster. For a single integer, an `atomic` beats both.
</details>

---

**6. Why must you not copy a `sync.Mutex` by value? What catches it?**

<details>
<summary>Answer</summary>

A mutex's protection lives in its internal state. Copying it (e.g. passing a
struct-with-mutex by value, or a value receiver) creates an independent lock;
goroutines then lock different copies and mutual exclusion is silently broken.
`go vet`'s **copylocks** check catches it. Fix: use pointer receivers and pass
pointers.
</details>

---

**7. What happens when multiple goroutines write a plain map concurrently?**

<details>
<summary>Answer</summary>

It's a data race *and* the Go runtime actively detects concurrent map writes and
calls `fatal error: concurrent map writes`, crashing the program — this detection
is always on, even without `-race`. Fix with a mutex/RWMutex around the map, or
`sync.Map` for specific read-mostly/disjoint-key workloads (benchmark first;
map+RWMutex frequently wins).
</details>

---

**8. When is `sync.RWMutex` actually better than `sync.Mutex`?**

<details>
<summary>Answer</summary>

When reads vastly outnumber writes **and** the critical section is non-trivial, so
many concurrent `RLock` readers proceed in parallel instead of serializing. For
tiny critical sections, `RWMutex`'s extra bookkeeping can make it *slower* than a
plain `Mutex` — so measure. A read-mostly config cache reloaded occasionally is the
canonical good fit.
</details>

---

**9. When should you use `atomic` instead of a mutex?**

<details>
<summary>Answer</summary>

For a single word-sized value with simple operations — a counter, a flag, a
swappable pointer. `atomic.Int64.Add(1)` / `Load()` make read-modify-write
indivisible without a lock, which is faster and can't deadlock. As soon as the
invariant spans **multiple** variables (they must change together), you need a
mutex — atomics only protect one value at a time.
</details>

---

**10. How does context cancellation actually stop a blocking operation?**

<details>
<summary>Answer</summary>

`ctx.Done()` returns a channel that's **closed** when the context is cancelled or
its deadline passes. Blocking code `select`s on it:

```go
select {
case res := <-work:   return res, nil
case <-ctx.Done():    return zero, ctx.Err() // Canceled or DeadlineExceeded
}
```

Closing `Done()` is a broadcast, so every goroutine waiting on that context wakes
at once. Library calls (`database/sql`, `net/http`) take a `ctx` and do this
internally, which is why every I/O call should accept a context.
</details>

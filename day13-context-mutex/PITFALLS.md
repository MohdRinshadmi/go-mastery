# Day 13 — Context, Mutexes, Race Detector: Pitfalls

Concurrency gotchas as **Trap → Why → Fix**.

---

### 1. Concurrent map access without a lock

**Trap:** Multiple goroutines write (or one writes while others read) a plain
`map` — `m[k]++`, `m[k] = v`.

**Why:** Go maps are not safe for concurrent writes. You get a data race *and* the
runtime may `fatal error: concurrent map writes` (a hard crash, detected even
without `-race`). `m[k]++` also loses updates because it's load-add-store.

**Fix:** Guard the map with a `sync.Mutex`/`RWMutex` next to it, or use
`sync.Map` for specific read-mostly cases (benchmark — map+RWMutex often wins).

---

### 2. Forgetting `defer cancel()`

**Trap:**
```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
result, err := doWork(ctx)
return result, err // cancel never called
```

**Why:** `WithCancel`/`WithTimeout`/`WithDeadline` allocate resources (a timer
and a goroutine tracking the parent). Not calling `cancel` leaks them until the
parent is cancelled — a slow goroutine/timer leak per call. `go vet`'s
`lostcancel` check flags it.

**Fix:** `defer cancel()` immediately after creating the context — even on the
happy path. Calling `cancel` twice is harmless.

---

### 3. Copying a `sync.Mutex` (or struct containing one) by value

**Trap:** `func update(c Counter)` where `Counter` embeds a `sync.Mutex`, or
returning a struct-with-mutex by value.

**Why:** Copying duplicates the lock's internal state. Goroutines then lock
*different* copies and the mutual exclusion is broken — silent races.

**Fix:** Use pointer receivers and pass `*Counter`. `go vet` copylocks catches
most cases. Never embed a mutex in a type you pass by value.

---

### 4. Storing a `context.Context` in a struct field

**Trap:** `type Server struct { ctx context.Context }`, then methods read `s.ctx`.

**Why:** Context is meant to flow *per call* down the stack, carrying that call's
deadline and cancellation. Stashing it in a struct ties one request's lifetime to
the struct and breaks per-call cancellation; it also invites use-after-cancel.

**Fix:** Pass `ctx` as the first parameter of each method: `func (s *Server) Do(ctx
context.Context, ...)`. Struct stores config, not request scope.

---

### 5. Early `return` with the lock held (no `defer`)

**Trap:**
```go
c.mu.Lock()
if cond { return }      // forgot to Unlock — lock held forever
c.n++
c.mu.Unlock()
```

**Why:** Any path that returns without unlocking leaves the mutex held → the next
`Lock()` deadlocks → the service hangs.

**Fix:** `defer c.mu.Unlock()` right after `Lock()`. It runs on every return path,
including panics.

---

### 6. Locking around slow I/O (coarse locking)

**Trap:** Holding a mutex across a network/DB call or other slow work.

**Why:** Every other goroutine blocks on the lock for the whole I/O duration —
you've serialized what should be concurrent and created a contention bottleneck
(visible in pprof as time in `sync.(*Mutex).Lock`).

**Fix:** Lock only the critical section that touches shared memory; do the slow
I/O outside the lock. Compute under the lock, call out unlocked.

---

### 7. Trusting "it works" without `-race`

**Trap:** Concurrent code passes tests and dev traffic, so you ship it.

**Why:** Data races are nondeterministic — they pass in dev and corrupt data
under production load. You cannot find them by reading code.

**Fix:** Run every package with goroutines under `go test -race` / `go run -race`
in CI. The detector reports the exact two racing stack traces. ~5–10× overhead, so
it's a test tool, not a production build flag.

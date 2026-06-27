# Challenge 03 — the counter that can't count

**Phase 3 · Concurrency · data races**

## Symptom

We fan out 1000 goroutines, each incrementing a shared counter once. We `Wait` for all of them, then print the total. It should be exactly `1000`.

It isn't. Run it a few times:

```bash
cd bugged
go run .
go run .
go run .
```

You'll see `987`, `1000`, `994`, `1000`... — non-deterministic and usually *less* than 1000. Counts are silently lost.

## Hint

Don't theorize — let the tool tell you:

```bash
go run -race .
```

Read the `WARNING: DATA RACE` block. It names the exact line doing an unsynchronized read *and* write from multiple goroutines. `counter++` is not one operation — it's load, add, store. Two goroutines interleaving those three steps lose increments.

## How to reproduce

`go run -race .` in `bugged/` — flagged every time. Without `-race` the wrong total appears intermittently; loop it a few times to see it.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

`counter++` performed concurrently from many goroutines with no synchronization is a **data race**. `counter++` compiles to three steps: read `counter`, add 1, write back. When goroutine A reads `5`, goroutine B also reads `5`, both compute `6`, both write `6` — two increments collapsed into one. The lost-update count is non-deterministic, which is why the total drifts below 1000 and changes run to run.

A data race is also *undefined behavior* in Go, not merely "wrong number" — the compiler and CPU are free to reorder and cache, so you can't reason about it at all. The race detector exists precisely because these bugs are invisible to the eye and intermittent in testing.

### The fix

Pick the right synchronization primitive for the job:

**Option A — `sync/atomic` (best for a single counter):**

```go
var counter atomic.Int64
// in each goroutine:
counter.Add(1)
// at the end:
fmt.Println(counter.Load())
```

Atomic operations make the read-add-write a single indivisible step. This is the fastest, lock-free option and exactly what counters are for.

**Option B — `sync.Mutex` (when you guard more than one field):**

```go
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()
```

`fixed/main.go` uses `atomic.Int64` because the workload is literally one counter — reach for atomics there, and a mutex once the invariant spans multiple variables.

Rule:

> Any variable read by one goroutine and written by another needs synchronization — a mutex, an atomic, or a channel that transfers ownership. "It's just an int" is not an exception.

Verify the fix is clean: `go run -race .` should print `1000` with no warning, every time.

</details>

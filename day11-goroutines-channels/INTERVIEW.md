# Day 11 — Goroutines & Channels: Interview Q&A

Model answers in `<details>`. Try to answer before expanding.

---

**1. What is the difference between a goroutine and an OS thread? How does the Go scheduler map between them?**

<details>
<summary>Answer</summary>

A goroutine is a lightweight, runtime-managed execution unit with a small
(~2 KB), dynamically growing stack; an OS thread has a large fixed stack
(~1–8 MB) and is scheduled by the kernel. Goroutines are far cheaper to create
(~1 µs vs ~10–100 µs) and switch (~100 ns user-space vs ~1–10 µs kernel-space),
so you can run millions of goroutines but only thousands of threads.

The runtime uses an **M:N scheduler**: M goroutines (**G**) are multiplexed onto
N OS threads (**M**) via P logical processors (**P**, = `GOMAXPROCS`). When a
goroutine blocks on a channel or network I/O, the runtime parks it and runs
another G on the same M, so the OS thread itself never blocks on user-land I/O.
That's why a Go server handles tens of thousands of connections on a handful of
threads.
</details>

---

**2. What does "Do not communicate by sharing memory; share memory by communicating" mean in practice? Give a code example.**

<details>
<summary>Answer</summary>

Instead of multiple goroutines locking a shared variable, give one goroutine
ownership of the data and have others send it messages over a channel. The
channel *is* the synchronization, so no lock is needed.

```go
// One owner goroutine holds the counter; others request increments.
type req struct{ resp chan int }
inc := make(chan req)
go func() {
    n := 0
    for r := range inc { n++; r.resp <- n }
}()
r := req{resp: make(chan int)}
inc <- r
fmt.Println(<-r.resp) // 1
```

This isn't a rule against mutexes (Day 13) — it's a default: prefer channels for
coordinating goroutines, reach for a mutex when you just guard a small piece of
shared state.
</details>

---

**3. What is an unbuffered channel? What happens when you send on one with no receiver ready?**

<details>
<summary>Answer</summary>

An unbuffered channel (`make(chan T)`) has zero capacity. A send blocks until a
receiver is ready, and a receive blocks until a sender sends — they meet in a
**rendezvous/handshake**. A send with no concurrent receiver blocks the sending
goroutine indefinitely. If that's the only goroutine, you get `fatal error: all
goroutines are asleep - deadlock!`; if it's one of many, it silently leaks.
</details>

---

**4. What is a goroutine leak? How do you detect one? How do you prevent one?**

<details>
<summary>Answer</summary>

A goroutine leak is a goroutine that was started but never exits — typically
blocked forever on a channel send/receive, or spinning with no stop condition.
It holds its stack and referenced memory (and maybe locks/FDs) for the life of
the process; per-request leaks accumulate into an OOM.

**Detect:** `runtime.NumGoroutine()` in tests (assert it returns to baseline),
`go.uber.org/goleak`, or a pprof goroutine profile in production.

**Prevent:** every goroutine must have a clear exit condition — the work finishes,
a context is cancelled (`<-ctx.Done()`), or a channel is closed. If you can't
state in one sentence what stops a goroutine, the design is wrong.
</details>

---

**5. What happens when you receive from a closed channel? What happens when you send to a closed channel?**

<details>
<summary>Answer</summary>

Receiving from a closed (and drained) channel returns immediately with the zero
value and `ok == false`: `v, ok := <-ch`. That's how `range ch` knows to stop.
Sending to a closed channel **panics** (`send on closed channel`). Closing an
already-closed channel also panics. Rule: only the sender closes, exactly once.
</details>

---

**6. Why do directional channels (`chan<-`, `<-chan`) exist? Give a real use case.**

<details>
<summary>Answer</summary>

They encode the producer/consumer contract in the type system. `chan<- T` is
send-only, `<-chan T` is receive-only; the compiler rejects misuse. A generator
returns `<-chan int` so callers can only read it, never send or close it (closing
is the generator's job). In a large codebase this prevents two teams from
accidentally wiring a producer to another producer — a bug that compiles fine
with bidirectional channels.
</details>

---

**7. Explain the classic loop-variable closure bug. Why does Go 1.22+ partially address it, and why must you still understand it?**

<details>
<summary>Answer</summary>

Pre-1.22, a `for` loop reused a single variable across iterations; goroutines
capturing it by reference all read the *final* value after the loop ended
(`5,5,5,5,5`). Go 1.22 made each iteration have its own copy of the loop
variable, fixing this exact pattern. You still must understand it because (a) you
read and maintain pre-1.22 code, and (b) the underlying hazard — capturing *any*
shared mutable variable in a goroutine — is unchanged. Passing the value as an
argument (`go func(i int){...}(i)`) is explicit and version-independent.
</details>

---

**8. Why is `make(chan T)` for two goroutines a synchronization point, not just data transfer?**

<details>
<summary>Answer</summary>

Because the receive can't complete until the send happens (and vice versa), the
receiving goroutine is guaranteed that the sending goroutine has executed up to
its send statement. The Go memory model formalizes this: a send on a channel
*happens-before* the corresponding receive completes. So an unbuffered channel
both moves a value *and* establishes ordering — you can use it purely for
signaling (`chan struct{}`) with no data at all.
</details>

---

**9. How do you wait for a single goroutine to finish without `sync.WaitGroup`?**

<details>
<summary>Answer</summary>

Use a `done` channel of `struct{}` (zero-size, signal-only):

```go
done := make(chan struct{})
go func() { defer close(done); work() }()
<-done // blocks until the goroutine closes done
```

`close(done)` is a broadcast — every receiver (here just one) unblocks. For N
goroutines a `WaitGroup` is cleaner (Day 12), but the channel form is the
primitive everything else is built on.
</details>

---

**10. You launch a goroutine in an HTTP handler that outlives the request. What's the risk and the fix?**

<details>
<summary>Answer</summary>

The goroutine may use a cancelled request context or a garbage-collected request
object, and it leaks if it blocks. If it must outlive the request, give it an
independent lifetime: a fresh context (e.g. `context.WithTimeout(context.
Background(), d)`), copies of the data it needs (not the request), and a clear
stop condition. Otherwise, do the work synchronously or hand it to a bounded
worker pool (Day 14) rather than firing-and-forgetting.
</details>

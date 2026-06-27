# Day 11 — Goroutines & Channels

> Mentor note: This is Go's crown jewel. Everything from the scheduler design to the "communicate, don't share" philosophy was a deliberate reaction to the pain of traditional concurrency. Read every "Senior take" box — this is where juniors write bugs that take down production at 2 a.m.

---

## 0. Go's Concurrency Philosophy

Before a single line of code: understand *why* Go's model is different.

In most languages (Java, Python, C++), shared-memory concurrency means:
1. Spin up threads.
2. Protect shared data with locks.
3. Hope you didn't deadlock, starve, or miss a lock.

Go's core insight, borrowed from Tony Hoare's **Communicating Sequential Processes (CSP)** paper (1978):

> "Do not communicate by sharing memory; instead, share memory by communicating."

What does that mean practically? Instead of a global counter that 10 goroutines all lock/unlock, you have one goroutine that *owns* the counter and other goroutines *send messages* to it asking for updates. The owner does the mutation; no lock needed. The channel *is* the synchronization.

This doesn't mean mutexes are wrong (Day 13). It means channels are the idiomatic Go tool when goroutines need to coordinate. Use each where it fits.

---

## 1. Goroutines

### Theory
A **goroutine** is a lightweight, cooperatively-scheduled function execution managed by the Go runtime. You launch one with the `go` keyword.

```go
go doSomething()     // runs doSomething() concurrently
go func() {          // anonymous goroutine
    fmt.Println("hello from goroutine")
}()
```

### Why goroutines exist (vs OS threads)

| | OS Thread | Goroutine |
|---|---|---|
| Stack size | ~1–8 MB fixed | ~2 KB, grows/shrinks dynamically |
| Cost to create | ~10–100 µs | ~1 µs |
| Scheduling | OS kernel (preemptive) | Go runtime (cooperative + preemptive since 1.14) |
| Context switch | ~1–10 µs, kernel space | ~100 ns, user space |
| Typical app limit | Thousands | Millions |

You can feasibly run **1,000,000 goroutines** in a Go program. You cannot run 1,000,000 OS threads. This is not an accident — it was the core design goal.

### The Go Scheduler (M:N model)

Go uses an **M:N scheduler**: M goroutines scheduled onto N OS threads, using P "processors" (GOMAXPROCS, defaults to # CPU cores).

```
  G  G  G  G  G      <- Goroutines (many)
  |  |  |  |  |
  P  P  P  P         <- Processors (GOMAXPROCS)
  |  |  |  |
  M  M  M  M         <- OS threads (M <= P typically)
  |  |  |  |
  OS KERNEL
```

When a goroutine blocks (e.g. network I/O, channel receive), the runtime parks it and runs another goroutine on the same thread. The OS thread never blocks on user-land I/O. This is why Go servers handle tens of thousands of concurrent connections with very few OS threads.

### When to use goroutines
- Any I/O-bound work: HTTP calls, DB queries, file reads — launch concurrently.
- CPU-bound work that can be parallelized (GOMAXPROCS > 1).
- Background tasks: timers, health checks, event loops.

### When NOT to launch a goroutine
- When the work is trivially fast and the overhead isn't worth it.
- When you have no plan to wait for it or communicate its errors — this is how goroutine leaks happen.

### Common mistake #1: fire-and-forget without cleanup

```go
// BAD: goroutine leaks if the work never finishes
func processRequest(r Request) {
    go doHeavyWork(r) // if this never returns, it runs forever
}
```

Every goroutine you launch should have a clear **lifetime**: it stops when the work is done, or when a context is cancelled, or when a channel is closed. No orphan goroutines.

### Common mistake #2: main() exits before goroutines finish

```go
func main() {
    go fmt.Println("hello") // might never print!
    // main exits — all goroutines killed
}
```

The Go runtime kills all goroutines when main() returns. You must **synchronize** — either with channels, WaitGroups (Day 12), or context (Day 13).

### Goroutine leak — a production horror story
A goroutine leak is a goroutine that was started but never exits. It holds memory, may hold locks, may hold file descriptors. In a server, every request that leaks one goroutine means slow death: the leak accumulates, memory grows, eventually the process OOMs and crashes.

**Detection:** Use `runtime.NumGoroutine()` in tests; use `goleak` library in tests; profile with `pprof` goroutine profile in production.

**Senior take:** In every function you write that launches goroutines, ask: "What stops this goroutine?" If you can't answer in one sentence, you have a design problem.

---

## 2. Channels

### Theory
A **channel** is a typed conduit for sending values between goroutines. It is the primary coordination primitive in Go.

```go
ch := make(chan int)      // unbuffered channel of int
ch <- 42                  // send 42 (blocks until receiver is ready)
v := <-ch                 // receive (blocks until sender is ready)
```

### Why channels exist
Channels provide **safe communication** between goroutines. No locks, no shared memory, no race conditions (on the channel values themselves — the channel is internally synchronized).

Under the hood a channel is: a ring buffer (for buffered), a mutex protecting it, and two wait queues (blocked senders, blocked receivers). You don't need to know this to use channels — but it explains the cost.

### Unbuffered channels = synchronization points

An unbuffered channel (`make(chan T)`) has **zero capacity**. A send blocks until a receiver is ready. A receive blocks until a sender sends. They meet in the middle — this is a **rendezvous** or **handshake**.

```go
ch := make(chan string)

go func() {
    ch <- "done" // blocks here until main receives
}()

msg := <-ch // blocks here until goroutine sends
fmt.Println(msg) // prints "done"
```

This is not just data transfer — it is **synchronization**. The receive guarantees the goroutine has run up to the send point.

### Closing a channel

```go
close(ch) // signals: no more values will be sent
```

Rules:
1. Only the **sender** closes a channel. Never the receiver.
2. Sending on a closed channel **panics**. Receiving from a closed channel returns the zero value immediately (and `ok=false`).
3. You can't re-open a closed channel.

```go
v, ok := <-ch
if !ok {
    // channel is closed and drained
}
```

### range over a channel

```go
for v := range ch { // receives until ch is closed
    fmt.Println(v)
}
```

This is the idiomatic way to consume a channel. The loop exits only when the channel is **closed and drained**. If you forget to close the channel, this loop hangs forever — a deadlock.

### Directional channels

You can restrict a channel to send-only or receive-only at the type level:

```go
func produce(out chan<- int) { // out: send-only
    out <- 42
}

func consume(in <-chan int) {  // in: receive-only
    v := <-in
    fmt.Println(v)
}
```

Why? **Documentation and safety.** If a function has `out chan<- int`, you know it's a producer. If it has `in <-chan int`, it's a consumer. The compiler enforces the contract — you can't accidentally receive from a producer's channel. Always annotate channel directions in function signatures.

### Common mistakes

1. **Deadlock from missing goroutine:**
   ```go
   ch := make(chan int)
   ch <- 1 // DEADLOCK: nobody receiving, main is blocked
   ```
   Unbuffered sends need a concurrent receiver. Always launch a goroutine or use a buffered channel.

2. **Forgetting to close — range hangs:**
   ```go
   ch := make(chan int)
   go func() {
       for i := 0; i < 3; i++ { ch <- i }
       // forgot: close(ch)
   }()
   for v := range ch { fmt.Println(v) } // DEADLOCK: waits forever
   ```

3. **Closing from the wrong end (multiple senders):**
   ```go
   // Two goroutines both calling close(ch) → PANIC
   go func() { close(ch) }()
   go func() { close(ch) }() // panic: close of closed channel
   ```
   With multiple senders, use a `sync.WaitGroup` to know when *all* senders are done, then have one coordinator goroutine close the channel. Or use `sync.Once`.

4. **nil channel blocks forever:**
   ```go
   var ch chan int // nil
   ch <- 1         // blocks forever (send on nil channel)
   <-ch            // blocks forever (receive on nil channel)
   ```
   Nil channels are useful in `select` (Day 12) to disable a case. Otherwise: always `make` your channels.

### Performance implications
- Channel operations involve at least one memory copy and possible scheduler involvement. For very high-throughput data (millions of ops/sec), a mutex-protected ring buffer can be faster. But for most code: channels are the right tool.
- An unbuffered channel is NOT free — every send/receive may cause a goroutine context switch. In a tight loop, this is measurable. Profile before optimizing.

---

## Beginner Example

```go
// Simplest possible usage: one goroutine sends, main receives.
ch := make(chan string)
go func() {
    ch <- "hello from goroutine"
}()
msg := <-ch
fmt.Println(msg)
```

## Production-grade Example: pipeline stage

```go
// generator creates a channel and feeds it from a goroutine.
// It returns a receive-only channel — callers can only read.
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out) // always close when done sending
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {   // range exits when in is closed
            out <- n * n
        }
    }()
    return out
}

// Usage:
// c := generate(2, 3, 4)
// for sq := range square(c) { fmt.Println(sq) } // 4 9 16
```

This is the **pipeline pattern** — each stage is a goroutine that reads from one channel and writes to another. We'll build full pipelines in Day 15.

---

## Expert Thinking Mode

- **Beginner:** "A goroutine is a thread. A channel is a queue between threads."
- **Senior:** "A goroutine is a concurrent execution unit with 2KB stack that the scheduler multiplexes. A channel isn't just a queue — it's a synchronization primitive. Closing a channel is a *broadcast*: every receiver wakes up. That's semantically different from a queue drain."
- **Staff:** "I design goroutine lifetimes as deliberately as I design function lifetimes. Every goroutine has an owner, an exit condition, and its errors are surfaced. I instrument goroutine counts in production because a goroutine leak is a memory leak is a crash."
- **Architect:** "The CSP model maps naturally to microservices: each service is a goroutine, each API call is a message on a channel. The same reasoning that makes intra-process communication safe makes inter-process design clean. I pick Go for systems where concurrency is the problem — not the afterthought."

---

## Real-world use

- **Kubernetes:** Uses goroutines heavily — one goroutine per informer watch, one per reconcile loop, one per API request. The controller-runtime library is essentially a carefully orchestrated set of goroutines and channels.
- **Nginx-replacement servers (Caddy, Traefik):** Handle each connection in a goroutine. The 2KB stack cost means you can have 100k concurrent connections in a few hundred MB of RAM.
- **Pub/Sub systems (NATS):** NATS server routes messages between subscribers using channels. The "message is a channel send" model maps directly.
- **etcd:** Uses channels for internal raft log processing — each log entry is sent on a channel to the commit loop.

---

## Common Race Conditions & Production Pitfalls

### Race: Closure captures loop variable by reference
```go
// CLASSIC BUG — ALL goroutines print the same (last) value
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // captures &i, not the value of i at launch
    }()
}
// Fix: pass i as argument or rebind: i := i
for i := 0; i < 5; i++ {
    i := i // shadow: each goroutine gets its own copy
    go func() {
        fmt.Println(i)
    }()
}
```

### Pitfall: goroutine launched in a request handler leaks after request ends
```go
// BAD: goroutine outlives the request, uses the request context after cancellation
func handler(ctx context.Context, req Request) {
    go func() {
        time.Sleep(10 * time.Second)
        // ctx may be cancelled, req may be garbage-collected
        doSomething(ctx, req)
    }()
}
```

### Pitfall: channel direction mismatch in a large codebase
Without directional channels, two teams can accidentally wire a producer to another producer. Use `chan<-` and `<-chan` in all function signatures as a compile-time contract.

---

## Interview Questions

1. What is the difference between a goroutine and an OS thread? How does the Go scheduler map between them?
2. What does "Do not communicate by sharing memory; share memory by communicating" mean in practice? Give a code example.
3. What is an unbuffered channel? What happens when you send on one with no receiver ready?
4. What is a goroutine leak? How do you detect one? How do you prevent one?
5. What happens when you receive from a closed channel? What happens when you send to a closed channel?
6. Why do directional channels (`chan<-`, `<-chan`) exist? Give a real use case.
7. Explain the classic loop-variable closure race condition. Why does Go 1.22+ partially address it, and why do you still need to understand it?

---

## Your tasks for today

Go to `../exercises/`. There are **3 beginner exercises** + **1 intermediate challenge**. Run `go run main.go` after each. No peeking at `../solutions/` until you've genuinely tried.

The goal: you should feel the difference between "goroutine launched" and "goroutine finished". That gap — that you can't feel it unless you synchronize — is the most important intuition in concurrent Go.

# Challenge 06 — the job queue that eats all the RAM

**Phase 6 · Advanced · backpressure & memory**

## Symptom

A job-processing service accepts work via `Submit(job)` and processes it with a pool of workers. Producers are fast; each job is a little slow to process. The intent is a bounded in-memory queue that applies **backpressure** — when the queue is full, `Submit` should block (or reject) so producers slow down to match the workers.

Instead, under a burst the resident memory climbs without limit. `Submit` never blocks no matter how far behind the workers fall, and the backlog of un-processed jobs grows until the process is OOM-killed. This program simulates a producer that vastly outpaces the workers and reports the peak backlog:

```bash
cd bugged
go run .
```

Expected: backlog stays bounded (≈ the queue capacity), `Submit` blocks once full.
Actual: the backlog balloons to the full producer count — no backpressure at all.

## Hint

How is the queue declared? `make(chan Job)` vs `make(chan Job, N)` vs *"append to a slice under a mutex"* behave very differently under load. A buffered channel of capacity `N` blocks the sender once `N` items are unconsumed — that block *is* your backpressure. A slice-backed queue that just keeps `append`-ing has no upper bound at all. Look at what `Submit` does when the consumers can't keep up — does it ever wait?

## How to reproduce

`go run .` in `bugged/`. The harness submits 50,000 jobs from a fast producer against a handful of slow workers and prints the maximum observed backlog.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The buggy queue is an unbounded slice guarded by a mutex:

```go
type Queue struct {
    mu   sync.Mutex
    jobs []Job
}

func (q *Queue) Submit(j Job) {
    q.mu.Lock()
    q.jobs = append(q.jobs, j) // grows without limit
    q.mu.Unlock()
}
```

`Submit` **always succeeds immediately** — `append` grows the slice however large it needs to. There is no upper bound and no mechanism to make a fast producer wait. When producers outrun consumers (the normal case under a traffic spike), the backlog and its memory grow without limit until the OOM killer steps in. The mutex makes it thread-safe but does nothing about *bounding* — thread-safe and bounded are different properties.

This is the canonical "unbounded queue" memory bug: people reach for a slice + mutex because it's familiar, and accidentally remove the one thing that gives a system backpressure.

### The fix

Use a **buffered channel** as the queue. Its capacity is a hard ceiling, and a send on a full channel blocks — which is exactly the backpressure you want:

```go
type Queue struct {
    jobs chan Job
}

func NewQueue(capacity, workers int) *Queue {
    q := &Queue{jobs: make(chan Job, capacity)}
    for i := 0; i < workers; i++ {
        go q.worker()
    }
    return q
}

func (q *Queue) Submit(j Job) {
    q.jobs <- j // blocks when the buffer is full -> producer slows down
}

func (q *Queue) worker() {
    for j := range q.jobs {
        process(j)
    }
}
```

Now memory is capped at `capacity` queued jobs (plus whatever the workers hold in flight). When the buffer fills, `Submit` blocks until a worker frees a slot, so producers naturally throttle to consumer speed. If blocking is unacceptable (e.g. an HTTP handler that must respond fast), use a non-blocking `select` to **shed load** instead:

```go
select {
case q.jobs <- j:
    // accepted
default:
    return ErrQueueFull // reject and let the caller retry/back off
}
```

Rules:

> 1. An in-memory queue must be **bounded**. A buffered channel gives you the bound and the backpressure in one primitive.
> 2. Decide up front: when full, do you **block** the producer or **shed** the load? Both are valid; an unbounded backlog is not.
> 3. "Thread-safe" (mutex) and "bounded" (capacity limit) are independent. You usually need both — a channel gives you both for free.

`fixed/` caps the backlog at the channel capacity; the bugged version lets it grow to the full submitted count.

</details>

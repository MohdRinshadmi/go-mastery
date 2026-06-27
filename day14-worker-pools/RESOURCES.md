# Day 14 — Worker Pools, Fan-Out/In, errgroup: Resources

- ★ **Go blog — "Go Concurrency Patterns: Pipelines and cancellation"** — bounded
  fan-out/fan-in and how to stop workers cleanly. Required reading for this day.
  https://go.dev/blog/pipelines

- ★ **`golang.org/x/sync/errgroup` docs** — the API for bounded, error-aware,
  context-cancelling concurrent groups (`Go`, `Wait`, `SetLimit`).
  https://pkg.go.dev/golang.org/x/sync/errgroup

- **Go by Example — Worker Pools** — minimal jobs/results pool you can run.
  https://gobyexample.com/worker-pools

- **Go by Example — WaitGroups** — the completion-tracking primitive behind the
  pool's `close(results)` coordinator.
  https://gobyexample.com/waitgroups

- **Rakyll — "errgroup" notes / x/sync overview** — practical errgroup usage and
  why it replaces hand-rolled error+cancel plumbing.
  https://pkg.go.dev/golang.org/x/sync

- **Bryan Mills — "Rethinking Classical Concurrency Patterns" (GopherCon talk)** —
  worker pools, semaphores, and the pitfalls of unbounded fan-out.
  https://www.youtube.com/watch?v=5zXAHh5tJqQ

- **`sync.WaitGroup` docs** — exact contract for the coordinator pattern.
  https://pkg.go.dev/sync#WaitGroup

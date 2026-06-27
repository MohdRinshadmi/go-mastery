# Day 12 — select, buffered channels, WaitGroup, Once: Resources

- ★ **Go blog — "Go Concurrency Patterns: Timing out, moving on"** — `select`
  with timeouts and the patterns this day is built on.
  https://go.dev/blog/concurrency-timeouts

- ★ **Effective Go — Channels & Parallelization** — buffered channels as
  semaphores, WaitGroup-style fan-out.
  https://go.dev/doc/effective_go#channels

- **`sync` package docs — WaitGroup, Once, Mutex** — the authoritative method
  contracts (Add-before-go, no-copy, exactly-once).
  https://pkg.go.dev/sync

- **`time` package — After, NewTimer, NewTicker** — why `After` leaks in loops
  and what to use instead.
  https://pkg.go.dev/time#After

- **Go spec — Select statements** — the exact semantics, including uniform random
  choice among ready cases.
  https://go.dev/ref/spec#Select_statements

- **Dave Cheney — "Curious Channels"** — nil channels, closed channels, and how
  they interact with `select`.
  https://dave.cheney.net/2013/04/30/curious-channels

- **`go vet` copylocks** — the check that catches copying a WaitGroup/Mutex.
  https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/copylock

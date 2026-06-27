# Day 11 — Goroutines & Channels: Resources

Curated, day-specific. Read the starred ones.

- ★ **Effective Go — Goroutines & Channels** — the canonical introduction.
  https://go.dev/doc/effective_go#goroutines

- ★ **Rob Pike — "Concurrency is not Parallelism" (talk)** — the mental model for
  why Go's concurrency is about *structure*, not just speed.
  https://go.dev/blog/waza-talk

- **The Go Memory Model** — what "happens-before" means; why a channel send
  synchronizes with its receive. Essential before you trust any goroutine code.
  https://go.dev/ref/mem

- **Go blog — "Share Memory By Communicating"** — the CSP philosophy with code.
  https://go.dev/blog/codelab-share

- **Go Tour — Goroutines & Channels** — interactive, run-in-browser exercises.
  https://go.dev/tour/concurrency/1

- **Go 1.22 release notes — loop variable change** — exactly what changed about
  per-iteration loop variables and why.
  https://go.dev/blog/loopvar-preview

- **Dave Cheney — "Never start a goroutine without knowing how it will stop"** —
  the leak-prevention discipline behind today's debugging exercise.
  https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop

- **uber-go/goleak** — the goroutine-leak detector to wire into tests.
  https://github.com/uber-go/goleak

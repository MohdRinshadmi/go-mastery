# Day 15 — Pipelines & Cancellation: Resources

- ★ **Go blog — "Go Concurrency Patterns: Pipelines and cancellation"** — THE
  reference for this day: stage shape, the done/ctx teardown, fan-out/in. Read it
  end to end.
  https://go.dev/blog/pipelines

- ★ **Rob Pike — "Concurrency is not Parallelism"** — the dataflow mindset behind
  composing stages.
  https://go.dev/blog/waza-talk

- **Go blog — "Go Concurrency Patterns: Context"** — how context replaces the raw
  `done` channel for cancellation across stages.
  https://go.dev/blog/context

- **uber-go/goleak** — the goroutine-leak detector for proving a pipeline cleans
  up after an early consumer exit.
  https://github.com/uber-go/goleak

- **The Go Memory Model** — why sending values (not sharing pointers) between
  stages is safe.
  https://go.dev/ref/mem

- **Sameer Ajmani — "Advanced Go Concurrency Patterns" (talk)** — `select`-driven
  state machines and cancellation that build on pipelines.
  https://go.dev/blog/io2013-talk-concurrency

- **"Concurrency in Go" (Katherine Cox-Buday) — pipelines chapter** — book-length
  treatment of the patterns and their failure modes.
  https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/

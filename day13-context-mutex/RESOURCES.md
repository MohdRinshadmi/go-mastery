# Day 13 — Context, Mutexes, Race Detector: Resources

- ★ **Go blog — "Go Concurrency Patterns: Context"** — the original context
  rationale and patterns; required reading.
  https://go.dev/blog/context

- ★ **The Go Memory Model** — what synchronization (mutex/channel/atomic) actually
  guarantees, and why an unsynchronized access is undefined.
  https://go.dev/ref/mem

- **Go blog — "Introducing the Go Race Detector"** — how `-race` works and what its
  output means.
  https://go.dev/blog/race-detector

- **Data Race Detector (reference)** — usage, options, and limitations.
  https://go.dev/doc/articles/race_detector

- **`context` package docs** — the authoritative API and usage rules.
  https://pkg.go.dev/context

- **`sync` package docs** — Mutex, RWMutex, Map contracts (including the
  no-copy-after-use rule).
  https://pkg.go.dev/sync

- **`sync/atomic` package docs** — the typed atomics (`atomic.Int64`, `Pointer`).
  https://pkg.go.dev/sync/atomic

- **Dave Cheney — "Context isn't for cancellation"** — a critical take on context
  scope and the Value anti-pattern.
  https://dave.cheney.net/2017/01/26/context-is-for-cancelation

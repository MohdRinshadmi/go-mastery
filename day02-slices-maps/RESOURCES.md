# Day 02 Resources — Slices & Maps

- **Go Slices: usage and internals (The Go Blog)** — https://go.dev/blog/slices-intro
  The canonical explanation of the slice header, capacity, and append behaviour.

- **Arrays, slices (and strings): The mechanics of 'append' (The Go Blog)** —
  https://go.dev/blog/slices
  Rob Pike's deep dive into how `append` and the backing array really work.

- **Go maps in action (The Go Blog)** — https://go.dev/blog/maps
  Idioms for maps: comma-ok, deletion, key types, iteration order.

- **Effective Go — Slices / Maps / Two-dimensional slices** —
  https://go.dev/doc/effective_go#slices
  Idiomatic usage straight from the official guide.

- **SliceTricks (Go wiki)** — https://go.dev/wiki/SliceTricks
  A cookbook of allocation-aware slice operations (delete, insert, filter, dedupe).

- **Go by Example: Slices / Maps** — https://gobyexample.com/slices ,
  https://gobyexample.com/maps
  Runnable snippets covering the everyday operations.

- **The Go spec — Appending to and copying slices** —
  https://go.dev/ref/spec#Appending_and_copying_slices
  The precise semantics of `append` and `copy`.

- **`maps` package (Go 1.21+)** — https://pkg.go.dev/maps
  `maps.Clone`, `maps.Keys`, `maps.Equal` and friends.

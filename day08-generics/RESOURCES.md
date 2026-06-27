# Day 08 — Generics Resources

Curated, real links. Start with the first two — they cover 90% of practical judgment.

- **[An Introduction to Generics](https://go.dev/blog/intro-generics)** — The official Go
  blog launch post; the clearest first read on type parameters, constraints, and `~`.

- **[When To Use Generics](https://go.dev/blog/when-generics)** — Ian Lance Taylor's
  guidance on the judgment call: data structures and same-behavior-many-types yes,
  reflexive `[T any]` no. The single most useful link here.

- **[Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)** —
  Hands-on, write-it-yourself walkthrough from the official docs.

- **[Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)** —
  The deep design document; read when you want the *why* behind type sets, `~`, and
  inference rules.

- **[`cmp` package docs](https://pkg.go.dev/cmp)** — Standard-library `cmp.Ordered`,
  `cmp.Compare`, `cmp.Less` — use these instead of `golang.org/x/exp/constraints`.

- **[`slices` package docs](https://pkg.go.dev/slices)** — Generic `Sort`, `Contains`,
  `Max`, `Index`, `BinarySearch` — the code you used to copy-paste, now in the stdlib.
  (Pair with **[`maps`](https://pkg.go.dev/maps)** for `Keys`, `Values`, `Clone`.)

- **[GopherCon 2021: Generics! — Robert Griesemer & Ian Lance Taylor](https://www.youtube.com/watch?v=Pa_e9EeCdy8)** —
  The designers walk through the feature and the tradeoffs; great for the mental model.

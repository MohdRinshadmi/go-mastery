# Day 07 — Resources

Curated, real links. Read the first three before the exercises; the rest deepen
specific patterns.

- [Effective Go — Embedding](https://go.dev/doc/effective_go#embedding) — the
  canonical explanation of promotion and why embedding is delegation, not
  inheritance, straight from the Go team.

- [The Go Blog — A GIF decoder / interfaces & composition](https://go.dev/blog/) —
  the official blog's interface and composition posts; small interfaces composed
  by consumers is the throughline.

- [Dave Cheney — SOLID Go Design](https://dave.cheney.net/2016/08/20/solid-go-design)
  — how SOLID maps onto Go: interface segregation, dependency inversion, and "the
  bigger the interface, the weaker the abstraction." Essential mindset.

- [Dave Cheney — Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
  — the original, definitive write-up of the functional options pattern and why it
  beats a config struct.

- [Rob Pike — Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html)
  — the post that inspired functional options; the closure-based options idea from
  the source.

- [Preslav Rachev — "Accept Interfaces, Return Structs" — Where Does the Idiom Come From?](https://preslav.me/2023/12/15/golang-accept-interfaces-return-structs/)
  — unpacks the mantra, when to follow it, and the rare cases to break it.

- [Refactoring Guru — Decorator pattern in Go](https://refactoring.guru/design-patterns/decorator/go/example)
  — a worked Go decorator example; pairs directly with the embed-the-interface
  middleware pattern in today's lesson.

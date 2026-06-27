# Day 06 Resources — Methods & Interfaces

Curated, all worth your time. Read the first three before anything else.

- [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interfaces_and_types)
  The canonical introduction: how methods attach to types and how interfaces are
  satisfied implicitly. Start here.

- [Go spec — Method sets](https://go.dev/ref/spec#Method_sets)
  The precise rule for what's in the method set of `T` vs `*T`. This is the
  authority behind every "type doesn't implement interface" error.

- [The Laws of Reflection (go.dev/blog)](https://go.dev/blog/laws-of-reflection)
  Explains the `(type, value)` representation of an interface value — the exact
  mental model that demystifies the nil-interface gotcha.

- [Dave Cheney — Understand Go interfaces and the nil panic / typed nil](https://dave.cheney.net/2017/08/09/typed-nils-in-go-1-9)
  The definitive write-up on typed nils and why `err != nil` lies. Pairs directly
  with today's debugging challenge.

- [Dave Cheney — How to use interfaces in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
  SOLID design in Go: small interfaces, accept interfaces / return structs, and
  defining interfaces in the consumer package.

- [Go Proverbs — Rob Pike](https://go-proverbs.github.io/)
  Short, quotable design wisdom: "The bigger the interface, the weaker the
  abstraction" and "Accept interfaces, return structs." Watch the linked talk.

- [Go blog — Errors are values / working with errors](https://go.dev/blog/errors-are-values)
  Reinforces treating `error` as an ordinary interface value, which is why typed
  nils and over-broad error handling go wrong.

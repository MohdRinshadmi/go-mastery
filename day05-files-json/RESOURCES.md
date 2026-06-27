# Day 05 Resources — Files, io, JSON

- **JSON and Go (The Go Blog)** — https://go.dev/blog/json
  The canonical guide to `Marshal`/`Unmarshal`, struct tags, and dynamic JSON.

- **`encoding/json` package docs** — https://pkg.go.dev/encoding/json
  Full reference for `Marshal`, `Unmarshal`, `Decoder`, `Encoder`, and tag syntax.

- **`io` package docs** — https://pkg.go.dev/io
  `Reader`, `Writer`, `Copy`, `ReadAll`, and the composition helpers.

- **`bufio` package docs** — https://pkg.go.dev/bufio
  `Scanner`, `Reader`, `Writer`, and the `Buffer`/`Flush` details.

- **`os` package docs** — https://pkg.go.dev/os
  `ReadFile`, `WriteFile`, `Open`, `OpenFile`, and the `O_*` flags + permissions.

- **Go by Example: Reading Files / Writing Files / JSON** —
  https://gobyexample.com/reading-files , https://gobyexample.com/writing-files ,
  https://gobyexample.com/json
  Runnable snippets for each.

- **The Laws of Reflection (The Go Blog)** — https://go.dev/blog/laws-of-reflection
  Background on why only exported fields are visible to `encoding/json`.

- **Effective Go — Interfaces and methods** —
  https://go.dev/doc/effective_go#interfaces_and_types
  Why small interfaces like `io.Reader`/`io.Writer` are so composable.

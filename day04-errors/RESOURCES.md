# Day 04 Resources — Error Handling

- **Working with Errors in Go 1.13 (The Go Blog)** — https://go.dev/blog/go1.13-errors
  The definitive guide to `%w` wrapping, `errors.Is`, and `errors.As`.

- **Error handling and Go (The Go Blog)** — https://go.dev/blog/error-handling-and-go
  The foundational post on errors as values and the error interface.

- **Effective Go — Errors** — https://go.dev/doc/effective_go#errors
  Idiomatic error construction and the panic/recover section.

- **`errors` package docs** — https://pkg.go.dev/errors
  Reference for `New`, `Is`, `As`, `Join`, and `Unwrap`.

- **Defer, Panic, and Recover (The Go Blog)** — https://go.dev/blog/defer-panic-and-recover
  How `panic`/`recover` and deferred closures interact (incl. named returns).

- **Don't just check errors, handle them gracefully (Dave Cheney)** —
  https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
  A widely-cited essay on wrapping, sentinels, and error design.

- **Go Code Review Comments — Error Strings / Handle Errors** —
  https://go.dev/wiki/CodeReviewComments#error-strings
  The style rules linters enforce (lowercase, no punctuation, don't discard).

- **Go by Example: Errors / Custom Errors / Panic / Recover** —
  https://gobyexample.com/errors , https://gobyexample.com/custom-errors ,
  https://gobyexample.com/panic , https://gobyexample.com/recover
  Runnable snippets for each.

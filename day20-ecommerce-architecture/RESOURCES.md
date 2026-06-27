# Day 20 — Clean Architecture Resources

Curated, real links. Read the layout standard and one or two of the "how I
structure services" posts — they're the closest thing the Go community has to a
consensus on this.

## Layout & architecture

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
  — the widely-referenced `cmd/`, `internal/`, layered-packages layout. Note it's
  a community convention, not an official standard; read the README's caveats.
- [Uncle Bob — The Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
  — the original essay: concentric layers, dependencies pointing inward. The
  theory our `domain → service → transport` structure implements.
- [Ben Johnson — Standard Package Layout](https://www.gobeyond.dev/standard-package-layout/)
  — a pragmatic Go take: organize by *domain* and keep dependencies (DB, HTTP)
  in their own subpackages with the domain at the root.

## Writing the HTTP services

- [Mat Ryer — How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)
  — `NewServer` constructor, dependency injection, thin handlers, `run()` for a
  testable `main`. The patterns behind our composition root and handler shape.
- [Alex Edwards — Let's Go Further](https://lets-go-further.alexedwards.net/)
  — book-length, production-grade Go API: layered structure, RBAC/permissions,
  consistent error responses, the works. The deep dive for this capstone.

## Go language mechanics

- [Go 1.4 release notes — Internal packages](https://go.dev/doc/go1.4#internalpackages)
  — where `internal/` was introduced and the exact import rule it enforces.
- [Go blog — Package names](https://go.dev/blog/package-names)
  — naming guidance that keeps `service`, `repository`, `domain` packages clear
  and avoids stutter (`service.ProductService`).
- [The Go Programming Language Specification](https://go.dev/ref/spec)
  — the reference for import paths, packages, and interface semantics underlying
  all of the above.

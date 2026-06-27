# Debugging Challenge — The Nil Dependency at the Composition Root

A clean-architecture service that compiles, runs, and serves every request —
but always with the *wrong* status. A `GET` for a product that clearly exists
comes back `500`, not `200`. The handler is fine. The service is fine. The repo
is fine. The wiring is not.

## Symptom

`GET /products?id=p1` should find the seeded product and return `200`. Instead
the bugged version prints:

```
=== bugged ===
GET /products?id=p1  -> status 500 (want 200)
body: error: product service has no repository: internal error
```

Every request fails the same way regardless of the id — the service never even
reaches the data. That uniform 500 is the tell: it's not a data problem, it's a
*construction* problem.

## Repro

Bugged (wrong status — wiring failure):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day20-ecommerce-architecture/debugging/bugged
go run .
```

Expected (buggy) output:

```
=== bugged ===
GET /products?id=p1  -> status 500 (want 200)
body: error: product service has no repository: internal error
```

Fixed (correct status):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day20-ecommerce-architecture/debugging/fixed
go run .
```

Expected (correct) output:

```
=== fixed ===
GET /products?id=p1      -> status 200 (want 200)
body: product: Fountain Pen
GET /products?id=nope    -> status 404 (want 404)
body: error: not found
```

## Hint

Read `main()` — the composition root — top to bottom and ask: *is every
dependency actually constructed and injected?* Look at the argument passed to
`NewService(...)`. The service stores whatever it's handed; it can't invent a
repository it was never given. Nothing above `main()` is wrong.

<details>
<summary>Solution & why</summary>

The composition root is the **only** place that knows the concrete wiring: it
builds the in-memory repository, injects it into the service, injects the
service into the handler, and starts serving. In the bugged version `main()`
skips the first step and passes `nil` where the repo should go:

```go
// BUG: forgot to construct/inject the repository
func main() {
    svc := NewService(nil)          // repo is nil
    h := &productHandler{svc: svc}
    // ...every request now fails: the service has no data source
}
```

The service guards against the nil repo and returns a domain error instead of
panicking, so the boundary's `statusFor()` maps it to `500`. (Without the guard
the first request would panic with a nil-pointer dereference inside
`s.repo.FindByID` — same root cause, louder failure.)

The fix is to wire the dependency at the composition root:

```go
// FIX: build the concrete repo and inject it
func main() {
    repo := newInMemoryRepo()       // construct the dependency
    svc := NewService(repo)         // inject it
    h := &productHandler{svc: svc}
    // ...now p1 -> 200, missing -> 404 via statusFor()
}
```

Notice the layers themselves never changed. `domain`, `repository`, `service`,
and `transport` are all correct in both versions. The bug lives *only* in the
one place responsible for assembly. That's the whole point of a composition
root: when wiring is centralized, wiring bugs are centralized too — you fix one
function, not a scattered hunt.

**Rules of thumb:**

- The **composition root** (`cmd/api/main.go`) is the *only* place that knows
  concrete types and wires them together. Layers below it depend on interfaces
  and must never construct their own dependencies.
- **Never leave an injected dependency nil.** If a constructor requires a repo,
  give it a real one. Consider failing fast in the constructor
  (`if repo == nil { panic("nil repo") }`) so wiring bugs blow up at startup,
  not on the first request.
- Map domain errors to HTTP status in exactly **one** `statusFor()` function so
  the error *type* drives the response and the mapping never drifts.
- A uniform 500 on every request — independent of input — almost always means a
  construction/wiring fault, not a data fault. Check the composition root first.

</details>

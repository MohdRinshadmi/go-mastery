# Day 20 — Clean Architecture Interview Questions

Model answers below. Read the question, answer out loud, then expand the
details. The "Senior take" lines mark what separates a passing answer from a
strong one.

### 1. Which way do dependencies point in clean architecture, and what may the domain import?

<details>
<summary>Answer</summary>

Dependencies point **inward**. Outer layers know about inner layers, never the
reverse: `transport → service → repository-interface → domain`. The composition
root (outermost) wires it all.

The **domain is innermost and imports nothing** from your own layers, and none
of the infrastructure packages — no `net/http`, no `database/sql`, no web
framework. It holds pure types (`User`, `Product`, `Order`), domain errors
(`ErrNotFound`, `ErrValidation`), and domain behavior. Standard-library
primitives like `time` or `errors` are fine; technology/transport packages are
not.

**Senior take:** "Dependencies point inward" is the whole rule restated as
"abstractions don't depend on details." The domain is the stable core that
nothing else can force to change.

</details>

### 2. Why must the service layer not import `net/http`? How do you test it then?

<details>
<summary>Answer</summary>

If the service imports `net/http` it's welded to one transport. You can no
longer call it from a CLI, a queue consumer, or a gRPC server, and you can't
unit-test it without constructing `http.Request`/`ResponseWriter`. The handler's
job is to convert HTTP into a plain function call; the service must not know HTTP
exists.

You test it by passing plain Go values and injecting an **in-memory repo** that
implements the repository interface — no DB, no HTTP, no network:

```go
func TestCreateProduct(t *testing.T) {
    repo := newMemRepo()
    svc := NewProductService(repo)
    _, err := svc.Create(context.Background(), "Pen", -5)
    if !errors.Is(err, domain.ErrValidation) { t.Fatal("want validation error") }
}
```

**Senior take:** The ease of that test is the *proof* the layering is right. If
testing a business rule requires `httptest`, the rule is in the wrong layer.

</details>

### 3. What is a composition root and why centralize wiring there?

<details>
<summary>Answer</summary>

The composition root is the single place — `cmd/api/main.go` — that constructs
concrete implementations and injects them: open the DB, build the repos, build
the services with those repos, build the handlers with those services, start the
server. It's the *only* place that knows concrete types.

Centralizing it means: (1) all wiring is visible in one file; (2) swapping
in-memory for Postgres is a one-line change; (3) lower layers depend only on
interfaces and never construct their own dependencies, so they stay testable;
(4) wiring bugs (like a nil repo) are localized to one function instead of
scattered `sql.Open` calls and globals.

**Senior take:** "Decisions about *what implementation* should live as close to
`main` as possible; the rest of the code should only know interfaces." That's
dependency injection done by hand — no framework required in Go.

</details>

### 4. What's the difference between authentication and authorization, and where does each live?

<details>
<summary>Answer</summary>

- **Authentication (authn) — who are you.** Verify credentials (password, token,
  JWT), establish identity. Done at the edge: a login endpoint issues a token;
  middleware validates the token on each request and puts the identity into the
  request `context`.
- **Authorization (authz) / RBAC — what may you do.** Given an identity, decide
  if it's allowed to perform this action. Lives either in a dedicated authz
  middleware (`RequireRole(admin)`) for coarse route gating, or in the service
  as a business rule for fine-grained rules (e.g. "a customer may read only
  their own order").

**Senior take:** Keep them separate and consistent. Identity goes into context
in middleware; permission decisions go in *one* place per rule. Scattering
`if role == "admin"` across handlers is exactly how privilege-escalation bugs
are born — one forgotten check and the door's open.

</details>

### 5. How do domain errors become HTTP status codes, and why map them in one place?

<details>
<summary>Answer</summary>

The service returns *domain* errors (`ErrNotFound`, `ErrForbidden`,
`ErrValidation`, `ErrUnauthorized`). The HTTP layer translates them in exactly
one function, using `errors.Is`:

```go
func statusFor(err error) int {
    switch {
    case errors.Is(err, domain.ErrNotFound):    return 404
    case errors.Is(err, domain.ErrForbidden):   return 403
    case errors.Is(err, domain.ErrValidation):  return 422
    case errors.Is(err, domain.ErrUnauthorized):return 401
    default:                                     return 500
    }
}
```

One place means the same error always yields the same status, and adding a new
error category is a one-line edit. Duplicating the mapping per handler causes
drift — the same error returning 404 here and 500 there.

**Senior take:** The error *type* drives the response. The service speaks
business vocabulary; the boundary owns the HTTP translation table. Don't let
status codes leak up into the service or scatter down across handlers.

</details>

### 6. What does `internal/` enforce, and how does it help layering?

<details>
<summary>Answer</summary>

A package under `internal/` can only be imported by code rooted at the parent of
that `internal/` directory. The Go compiler *enforces* this — code in another
module (or outside the subtree) physically cannot import your
`internal/service`, `internal/repository`, etc.

It turns layering from a convention into a compiler-enforced boundary. Outsiders
can't reach past your public API into your internals, so you can refactor
internal packages freely, and you can't accidentally create cross-module
dependencies on what should be private.

**Senior take:** `internal/` is "private" for packages. Your layering is
enforced, not just hoped for — the compiler rejects the import rather than a
reviewer catching it later.

</details>

### 7. When is this layering overkill, and when is it essential?

<details>
<summary>Answer</summary>

- **Overkill** for a 200-line tool, a one-off script, a prototype, or a lambda
  with a single endpoint. A single `main.go` with handlers calling functions is
  fine and faster to write. Premature layering adds ceremony with no payoff.
- **Essential** when multiple engineers work in parallel, the app will live for
  years, you need to swap storage (in-memory ↔ Postgres), you must unit-test
  business rules without infrastructure, or the codebase will grow past a few
  thousand lines.

**Senior take:** Match structure to lifespan and team size. The cost of layering
is up-front ceremony; the cost of *not* layering is a 5000-line `main.go` no one
can change safely. Start simple, refactor toward layers when the seams start to
hurt — but know the target shape so the refactor is cheap.

</details>

### 8. Dependency inversion vs dependency injection — what's the difference?

<details>
<summary>Answer</summary>

They're related but distinct:

- **Dependency inversion (a principle):** high-level policy should not depend on
  low-level details; both depend on an abstraction. Here, the service depends on
  a `ProductRepo` *interface* that the service layer *owns* — not on the Postgres
  repo. The detail (Postgres) depends on the abstraction, inverting the usual
  arrow.
- **Dependency injection (a technique):** how a component *receives* its
  dependencies — passed in (constructor args) rather than constructed internally.
  `NewProductService(repo)` injects the repo.

Inversion is *what* the relationship is (depend on an interface you own).
Injection is *how* the concrete value gets supplied (passed in at the composition
root).

**Senior take:** In Go you get inversion by defining the interface in the
*consumer* package and injection by hand at `main`. No container, no annotations
— just constructors taking interfaces.

</details>

### 9. Where do transaction boundaries belong in this layering?

<details>
<summary>Answer</summary>

A transaction is a *use-case/business* concern — "reserve stock AND create the
order atomically" — so the boundary belongs at the **service (use-case) level**,
not in individual repo methods (too fine) and never in the handler (wrong layer).

The cleanest Go approach without leaking `*sql.Tx` into the service is a
unit-of-work / `WithinTx` helper the repo layer exposes:

```go
err := s.uow.WithinTx(ctx, func(repos RepoSet) error {
    if err := repos.Stock.Reserve(ctx, sku, qty); err != nil { return err }
    return repos.Orders.Save(ctx, order)
})
```

The service expresses "do these together"; the repo layer owns how the tx is
opened/committed/rolled back. The service still never imports `database/sql`.

**Senior take:** One service method = one transaction is a good default.
Crossing multiple aggregates atomically is a smell that your boundaries (or your
consistency requirements) need rethinking — possibly eventual consistency.

</details>

### 10. How does this modular monolith become microservices later?

<details>
<summary>Answer</summary>

The package boundaries you draw today (users vs catalog vs orders — *bounded
contexts*) are the service boundaries of tomorrow. The migration path:

1. Keep each context behind a clean service interface, communicating through
   those interfaces rather than reaching into each other's repos/tables.
2. Enforce the boundaries with `internal/` and per-context packages so coupling
   can't sneak in.
3. When a context needs independent scaling/deploy, replace its in-process
   service call with a network call (gRPC/HTTP) — the interface stays, the
   transport changes. The repo behind it gets its own datastore.
4. Split the data last; shared tables are the hardest coupling to undo.

**Senior take:** "Module boundaries today are service boundaries tomorrow."
Don't start with microservices — start with a *modular* monolith whose seams are
already in the right places, and extract a service only when there's a real
reason (independent scaling, team ownership, deploy cadence). Distribution is a
cost, not a feature.

</details>

### 11. How do you keep handlers thin — and how do you know when one isn't?

<details>
<summary>Answer</summary>

A thin handler does exactly four things: **decode** the request, **call** one
service method, **map** the error with `statusFor`, **encode** the response.
Anything else — validation, authz logic, pricing, orchestration, SQL — belongs
below it.

```go
func (h *productHandler) get(w http.ResponseWriter, r *http.Request) {
    p, err := h.svc.Get(r.Context(), r.PathValue("id"))
    if err != nil { writeErr(w, statusFor(err)); return }
    writeJSON(w, 200, p)
}
```

Smells that a handler is too fat: it contains `if`-chains of business rules, it
imports `database/sql`, it has more than a trivial amount of logic between decode
and encode, or you can't test the rule it enforces without `httptest`.

**Senior take:** A handler is a *translator*, not a place where decisions are
made. If you find yourself unit-testing a handler to cover a business rule, move
the rule into the service and test it there instead.

</details>

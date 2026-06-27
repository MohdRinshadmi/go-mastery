# Day 20 — Clean Architecture: The E-Commerce API Capstone

> Mentor note: This is where Phase 4 comes together. Everything you've learned — interfaces, DI, errors, repositories, handlers, middleware — assembles into one coherent service with a layout you'll recognize at every serious Go shop. The lesson today isn't a new language feature; it's **how to organize a codebase so a team can work in it for years without it rotting.** The runnable app in this folder is a real (if compact) e-commerce backend. Read it top to bottom, then extend it in the exercises.

---

## 1. The layered architecture

Dependencies point **inward**. Outer layers know about inner layers, never the reverse.

```
            cmd/api/main.go          ← composition root: wire everything
                  │ builds & injects
   ┌──────────────┼─────────────────────────────┐
   ▼              ▼                               ▼
transport/http   service (business logic)   repository (data)
 (handlers,        depends on repo            in-memory now,
  middleware,      INTERFACES                 Postgres later
  routing)              │                          ▲
        │               └── interface owned here ──┘
        ▼
     domain (User, Product, Order, domain errors)  ← innermost, depends on nothing
```

- **domain/** — pure types and domain errors. No imports of http, sql, etc. The core.
- **repository/** — data access behind interfaces *defined by the service*. Swappable (in-memory ↔ Postgres) without touching anything above.
- **service/** — business logic: validation, authorization rules, orchestration. Depends on repo *interfaces*, not implementations. This is where the rules live.
- **transport/http/** — translates HTTP ↔ service calls. Thin: decode, call service, map errors to status codes, encode. No business logic here.
- **cmd/api/main.go** — the **composition root**: constructs concrete repos, injects them into services, injects services into handlers, starts the server. The *only* place that knows the concrete wiring.

**Senior take:** The golden rule — **business logic never imports `net/http` or `database/sql`.** If your service layer imports the http package, the layering is broken and you can't test or reuse it. The handler converts HTTP to a plain function call; the service doesn't know HTTP exists.

## 2. Why this structure

- **Testable**: services tested with in-memory repos, no DB, no HTTP (Day 9).
- **Swappable**: move from in-memory to Postgres by changing one line in `main.go`.
- **Parallel work**: one engineer on handlers, another on services, another on repos — clear seams.
- **`internal/`**: the compiler forbids other modules from importing your `internal/` packages — your layering is enforced, not just hoped for.

## 3. Authentication & Role-Based Access Control (RBAC)

- **Authentication** (who are you): login verifies credentials, issues a token (JWT in Day 17; here a simple signed/opaque token for clarity). Middleware validates it on protected routes and puts the user identity in the request `context`.
- **Authorization / RBAC** (what may you do): users have roles (`customer`, `admin`). Middleware or the service checks the role before an action — e.g. only `admin` may create products.

```
POST /login         → returns a token
POST /products      → requires admin role        (RBAC)
GET  /products      → public
POST /orders        → requires any logged-in user
```

**Senior take:** Authentication and authorization are different. Authn proves identity; authz decides permission. Put identity into the context in middleware; enforce role rules in the service (business rule) or a dedicated authz middleware — but be consistent. Scattering `if role == "admin"` across handlers is how privilege-escalation bugs are born.

## 4. Error mapping at the boundary

The service returns *domain* errors (`ErrNotFound`, `ErrForbidden`, `ErrValidation`). The HTTP layer maps them to status codes in one place:

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
This is Day 4 (error taxonomy) + Day 18 (status codes) realized: the error *type* drives the HTTP response, and the mapping lives in exactly one function.

## Common mistakes
1. Business logic in handlers (validation, authz, orchestration) → untestable, duplicated.
2. Services importing `net/http`/`database/sql` → broken layering.
3. Repository returning driver errors instead of domain errors.
4. No composition root — `sql.Open` called in random packages, globals everywhere.
5. RBAC checks scattered and inconsistent.
6. Anemic domain with all logic in one giant handler ("transaction script") — fine for tiny apps, unmaintainable as it grows.

## Performance / operability
- The layering adds near-zero runtime cost (interface calls are cheap); the win is maintainability.
- Pass `ctx` from handler → service → repo so timeouts/cancellation flow through.
- Keep handlers thin so the hot path is service + repo, easy to profile (Day 10).

---

## Expert Thinking Mode — "build the backend"

- **Beginner:** "One `main.go`, handlers with SQL inside. It works!" (Until it's 5000 lines.)
- **Senior:** "Layered: domain/service/repo/transport, DI at a composition root, repo interfaces, domain errors mapped to status codes, in-memory repos for tests."
- **Staff:** "Where are transaction boundaries? Idempotency for order/payment? How do auth and RBAC compose? Is the domain model rich enough or leaking into services? Migration + zero-downtime deploy story."
- **Architect:** "Module boundaries today are service boundaries tomorrow. Is this a modular monolith that can split into services (Phase 6)? Bounded contexts (users vs catalog vs orders) drive package and eventual service decomposition."

---

## Real-world use

- **Standard Go layout** (`cmd/`, `internal/`, layered packages) is used at Uber, Stripe, and across the ecosystem; see `github.com/golang-standards/project-layout`.
- **Repository + service + handler** is the default shape of a Go web backend.
- **RBAC middleware + context identity** is how every production API enforces permissions.
- The modular-monolith-to-microservices path (Phase 6) starts from exactly this layering.

---

## Interview Questions

1. Which way do dependencies point in clean architecture? What may the domain import?
2. Why must the service layer not import `net/http`? How do you test it then?
3. What is a composition root and why centralize wiring there?
4. Difference between authentication and authorization? Where does each live?
5. How do domain errors become HTTP status codes, and why map in one place?
6. What does `internal/` enforce and how does it help layering?
7. When is this layering overkill, and when is it essential?

---

## The runnable app (this folder)

```
cmd/api/main.go                 composition root (wire + serve)
internal/domain/                User, Product, Order, Role, domain errors
internal/repository/            in-memory repos (swap for Postgres from Day 19)
internal/service/               auth, user, product, order business logic + RBAC
internal/transport/http/        handlers, auth middleware, router, error mapping
```

Run it:
```bash
go run ./cmd/api
# in another terminal:
curl -s -XPOST localhost:8080/register -d '{"email":"a@x.com","password":"pw","name":"Ada"}'
curl -s -XPOST localhost:8080/login    -d '{"email":"a@x.com","password":"pw"}'   # -> {"token":"..."}
TOKEN=...   # admin seeded as admin@shop.com / admin
curl -s -XPOST localhost:8080/products -H "Authorization: Bearer $TOKEN" -d '{"name":"Pen","price":2.5}'
curl -s localhost:8080/products
```

## Your tasks (`exercises/`)

`exercises/TASKS.md` lists extensions to implement on top of the runnable app: (1) add `GET /orders/{id}` with an ownership check (a customer may only read their own orders — RBAC at the service layer), (2) add stock tracking to products and reject orders that exceed stock with `ErrValidation`, (3) challenge: add an `admin`-only `GET /users` endpoint. Implement them in the layered structure, keep handlers thin, and bring it for a full PR-style review. Passing this completes Phase 4.

## Day 20 companion files

- [Debugging challenge](../debugging/README.md) — the nil-dependency-at-the-composition-root bug (wrong status from a wiring fault).
- [Pitfalls](../PITFALLS.md) — Trap → Why → Fix for the layering traps (logic in handlers, broken layering, leaked driver errors, nil DI, scattered RBAC, duplicated error mapping, anemic domain).
- [Interview questions](../INTERVIEW.md) — the lesson's 7 plus deeper ones (dependency inversion vs injection, transaction boundaries, monolith → microservices, thin handlers), with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference: layer diagram, the golden rule, composition root, `statusFor()`, authn/authz/RBAC, `internal/`.
- [Resources](../RESOURCES.md) — curated links on Go project layout, clean architecture, and writing HTTP services.

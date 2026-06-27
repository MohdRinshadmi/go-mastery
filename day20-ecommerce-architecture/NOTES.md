# Day 20 — Clean Architecture Cheatsheet

Quick reference for the layered e-commerce capstone. The structure isn't a
language feature — it's how you keep a codebase alive for years.

## The layer diagram (dependencies point inward)

```
                cmd/api/main.go            ← composition root: wire & inject
                      │ builds & injects
   ┌──────────────────┼──────────────────────────────┐
   ▼                  ▼                                ▼
transport/http     service (business logic)      repository (data)
 handlers,           depends on repo                in-memory now,
 middleware,         INTERFACES ──────────┐         Postgres later
 routing                                  │              ▲
   │                                      └─ interface ──┘
   ▼                                         owned by service
 domain (User, Product, Order, domain errors)  ← innermost, imports nothing
```

- **domain/** — pure types + domain errors. No `net/http`, no `database/sql`.
- **repository/** — data access behind interfaces *defined by the service*.
  Swappable in-memory ↔ Postgres.
- **service/** — validation, authz rules, orchestration. Depends on repo
  *interfaces*, not implementations.
- **transport/http/** — thin: decode → call service → `statusFor(err)` → encode.
- **cmd/api/main.go** — composition root; the only place that knows concrete wiring.

## The golden rule

> **Business logic never imports `net/http` or `database/sql`.**

If your service imports the http package, the layering is broken — you can't
test or reuse it. The handler converts HTTP into a plain function call; the
service doesn't know HTTP exists. Same for SQL: the repo hides it behind an
interface.

## Composition root responsibilities (`cmd/api/main.go`)

1. Open resources (DB, config) — the *only* place `sql.Open` is called.
2. Construct concrete repos: `repo := repository.NewPostgresProductRepo(db)`.
3. Inject repos into services: `svc := service.NewProductService(repo)`.
4. Inject services into handlers: `h := transport.NewProductHandler(svc)`.
5. Build the router + middleware and `ListenAndServe`.
6. Never leave an injected dependency nil — fail fast if a constructor needs one.

```go
func main() {
    db   := mustOpen(cfg.DSN)
    repo := repository.NewPostgresProductRepo(db)   // concrete
    svc  := service.NewProductService(repo)         // inject interface
    h    := transport.NewProductHandler(svc)
    log.Fatal(http.ListenAndServe(":8080", routes(h)))
}
```

## Error mapping: one `statusFor()` at the boundary

The service returns *domain* errors; the HTTP layer maps them in one function:

```go
func statusFor(err error) int {
    switch {
    case err == nil:                             return 200
    case errors.Is(err, domain.ErrNotFound):     return 404
    case errors.Is(err, domain.ErrForbidden):    return 403
    case errors.Is(err, domain.ErrValidation):   return 422
    case errors.Is(err, domain.ErrUnauthorized): return 401
    default:                                      return 500
    }
}
// every handler: w.WriteHeader(statusFor(err))
```

The error *type* drives the response, and the table lives in exactly one place.

## Authn vs authz / RBAC placement

| Concern | Question | Where |
|---|---|---|
| Authentication | Who are you? | Login endpoint issues token; middleware validates it and puts identity into `context`. |
| Authorization / RBAC | What may you do? | `RequireRole(admin)` middleware for coarse route gating; service-level check for fine-grained rules (own-resource ownership). |

Rules of placement:
- Put identity into `context` once, in middleware.
- Enforce each permission rule in *one* place — never copy-paste `if role ==
  "admin"` across handlers (privilege-escalation breeding ground).

```
POST /login      → issue token              (authn)
GET  /products   → public
POST /products   → RequireRole(admin)       (authz/RBAC)
POST /orders     → any authenticated user   (authn)
```

## `internal/` enforcement

Packages under `internal/` can only be imported by code rooted at `internal/`'s
parent. The compiler rejects outside imports — your layering becomes a
*compile-time* boundary, not a convention. Use it to keep service/repo/domain
private to your module.

## Quick checklist

- [ ] Domain imports nothing technological.
- [ ] Service imports repo *interfaces*, never `net/http`/`database/sql`.
- [ ] Repo translates driver errors → domain errors.
- [ ] Exactly one composition root; no `sql.Open` elsewhere; no globals.
- [ ] No injected dependency left nil.
- [ ] Handlers are thin: decode → call → `statusFor` → encode.
- [ ] RBAC enforced consistently in one place per rule.

## Key terms

- **Clean / layered architecture** — organizing code into concentric layers
  (domain → service → repo/transport) where dependencies point inward.
- **Domain** — the innermost layer: pure types, domain behavior, and domain
  errors; imports no infrastructure.
- **Composition root** — the single place (`cmd/api/main.go`) that constructs
  concrete types and injects them; the only thing that knows the wiring.
- **Dependency inversion** — high-level policy and low-level details both depend
  on an abstraction the consumer owns (the repo interface), not on each other.
- **Dependency injection** — supplying a component's dependencies from outside
  (constructor args) instead of constructing them internally.
- **Repository interface** — data-access contract defined by the service and
  implemented by storage adapters; lets you swap in-memory ↔ Postgres.
- **Authn vs authz** — authentication proves identity; authorization decides
  permitted actions. Distinct concerns, distinct placement.
- **RBAC** — role-based access control: permissions attached to roles
  (`customer`, `admin`), checked consistently before an action.
- **`internal/`** — Go's compiler-enforced package privacy; only the parent
  subtree may import it.
- **Error mapping / `statusFor()`** — the single function translating domain
  errors into HTTP status codes at the boundary.

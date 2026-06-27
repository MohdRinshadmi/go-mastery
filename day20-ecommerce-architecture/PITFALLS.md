# Day 20 — Clean Architecture Pitfalls (Trap → Why → Fix)

Layering bugs don't crash the compiler — they rot the codebase. Each pitfall
below compiles and "works" today; each one costs you for years.

## 1. Business logic living in the handler

**Trap.**

```go
func (h *orderHandler) create(w http.ResponseWriter, r *http.Request) {
    var in createOrderReq
    json.NewDecoder(r.Body).Decode(&in)
    if in.Qty <= 0 {                       // validation in the handler
        http.Error(w, "bad qty", 400)
        return
    }
    if r.Header.Get("X-Role") != "customer" { // authz in the handler
        http.Error(w, "forbidden", 403)
        return
    }
    total := in.Qty * priceOf(in.SKU)      // pricing rules in the handler
    // ...write to DB straight from the handler...
}
```

**Why.** Validation, authorization, and pricing are *business rules*. Buried in
a handler they're untestable without HTTP, impossible to reuse from a CLI or
worker, and silently duplicated the next time someone adds an endpoint.

**Fix.** Handlers only decode, call one service method, map the error, encode.

```go
func (h *orderHandler) create(w http.ResponseWriter, r *http.Request) {
    var in createOrderReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        w.WriteHeader(400); return
    }
    order, err := h.svc.CreateOrder(r.Context(), userFrom(r.Context()), in.SKU, in.Qty)
    if err != nil { w.WriteHeader(statusFor(err)); return }
    writeJSON(w, 201, order)
}
// All validation/authz/pricing rules live in svc.CreateOrder, tested without HTTP.
```

## 2. The service importing `net/http` or `database/sql`

**Trap.**

```go
package service

import (
    "net/http"
    "database/sql"
)

func (s *ProductService) Create(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
    // service now knows about requests and SQL drivers
}
```

**Why.** This is the cardinal sin. Once business logic imports `net/http` you
can't call it from anywhere but an HTTP handler, can't unit-test it without
spinning up requests, and can't swap the transport (gRPC, CLI, queue). Importing
`database/sql` couples your rules to one storage technology forever.

**Fix.** The service takes plain Go values and depends on repo *interfaces*. The
golden rule: **business logic never imports `net/http` or `database/sql`.**

```go
package service

type ProductRepo interface {                       // interface, not *sql.DB
    Save(ctx context.Context, p domain.Product) error
}
func (s *ProductService) Create(ctx context.Context, name string, price int) (domain.Product, error) {
    if price < 0 { return domain.Product{}, domain.ErrValidation }
    p := domain.Product{Name: name, Price: price}
    return p, s.repo.Save(ctx, p)
}
```

## 3. Repository leaking driver errors instead of domain errors

**Trap.**

```go
func (r *pgProductRepo) FindByID(ctx context.Context, id string) (domain.Product, error) {
    var p domain.Product
    err := r.db.QueryRowContext(ctx, "...").Scan(&p.ID, &p.Name)
    return p, err        // returns sql.ErrNoRows straight up
}
```

**Why.** Now the service (and worse, the handler) must `errors.Is(err,
sql.ErrNoRows)` — coupling upper layers to the SQL driver. Switch to Mongo or an
HTTP backend and every caller breaks. The abstraction has leaked.

**Fix.** Translate storage errors to domain errors at the repo boundary.

```go
func (r *pgProductRepo) FindByID(ctx context.Context, id string) (domain.Product, error) {
    var p domain.Product
    err := r.db.QueryRowContext(ctx, "...").Scan(&p.ID, &p.Name)
    if errors.Is(err, sql.ErrNoRows) {
        return domain.Product{}, domain.ErrNotFound   // domain vocabulary
    }
    return p, err
}
```

## 4. No composition root — `sql.Open` scattered with globals

**Trap.**

```go
package repository

var db, _ = sql.Open("postgres", os.Getenv("DSN")) // package-level global, opened on import
```

**Why.** Connections open as a side effect of importing a package. You can't
inject a test double, can't control startup order, can't see all wiring in one
place, and tests share hidden global state. Lifecycle is uncontrollable.

**Fix.** One composition root constructs everything and injects it down.

```go
func main() { // cmd/api/main.go — the ONLY place that opens resources
    db := mustOpen(os.Getenv("DSN"))
    repo := repository.NewPostgresProductRepo(db)
    svc  := service.NewProductService(repo)
    h    := transport.NewProductHandler(svc)
    http.ListenAndServe(":8080", routes(h))
}
```

## 5. A nil dependency injected (and never noticed)

**Trap.**

```go
svc := service.NewProductService(nil) // forgot the repo
h   := transport.NewProductHandler(svc)
// first request: nil-pointer panic deep in the service, or a silent 500
```

**Why.** Constructors that accept an interface happily store `nil`. The failure
surfaces far from the cause — a panic inside `svc.Get` or a uniform 500 — making
it look like a logic bug when it's a wiring bug. (See `debugging/`.)

**Fix.** Always inject a real dependency, and fail fast on nil at construction.

```go
func NewProductService(repo ProductRepo) *ProductService {
    if repo == nil { panic("NewProductService: nil repo") } // crash at startup, not in prod traffic
    return &ProductService{repo: repo}
}
```

## 6. RBAC checks scattered as `if role == "admin"` everywhere

**Trap.**

```go
func (h *productHandler) create(w, r) { if role(r) != "admin" { /*403*/ } /*...*/ }
func (h *userHandler) list(w, r)      { if role(r) != "admin" { /*403*/ } /*...*/ }
func (h *reportHandler) export(w, r)  { if role(r) == "admin" || role(r) == "ops" { /*...*/ } } // drift!
```

**Why.** String comparisons copy-pasted across handlers drift apart (note the
`ops` exception above). One missed check is a privilege-escalation bug. Authz
logic doesn't belong in the transport layer at all.

**Fix.** Centralize authorization: a `RequireRole` middleware and/or an authz
check in the service, expressed once.

```go
mux.Handle("POST /products", RequireRole(domain.RoleAdmin, h.create)) // one consistent gate
// or, as a business rule in the service:
func (s *ProductService) Create(ctx context.Context, actor domain.User, ...) error {
    if !actor.Can(domain.PermCreateProduct) { return domain.ErrForbidden }
    // ...
}
```

## 7. Error→status mapping duplicated instead of one `statusFor()`

**Trap.**

```go
// in handler A
if errors.Is(err, domain.ErrNotFound) { w.WriteHeader(404) } else { w.WriteHeader(500) }
// in handler B (forgot ErrForbidden, mapped validation to 400 not 422)
if errors.Is(err, domain.ErrNotFound) { w.WriteHeader(404) } else { w.WriteHeader(400) }
```

**Why.** Each handler invents its own mapping, so the same domain error returns
different status codes across endpoints. Adding `ErrForbidden` means hunting
every handler — and you'll miss one.

**Fix.** One `statusFor(err)` function owns the entire error→status table.

```go
func statusFor(err error) int {
    switch {
    case err == nil:                            return 200
    case errors.Is(err, domain.ErrNotFound):    return 404
    case errors.Is(err, domain.ErrForbidden):   return 403
    case errors.Is(err, domain.ErrValidation):  return 422
    case errors.Is(err, domain.ErrUnauthorized):return 401
    default:                                     return 500
    }
}
// every handler: w.WriteHeader(statusFor(err))
```

## 8. Anemic domain — the giant transaction-script handler

**Trap.**

```go
func checkout(w http.ResponseWriter, r *http.Request) {
    // 300 lines: decode, validate, price, apply discounts, reserve stock,
    // charge card, write order, send email — all inline, domain types are
    // just bags of public fields with no behavior
}
```

**Why.** Fine for a toy; a tar pit as it grows. No reuse, no unit tests, every
rule entangled with HTTP and SQL. The domain is "anemic" — data with no
behavior — so all logic congeals into one untestable handler.

**Fix.** Give the domain behavior and split orchestration into the service.

```go
// domain owns invariants
func (o *Order) AddLine(p Product, qty int) error {
    if qty <= 0 { return ErrValidation }
    o.Lines = append(o.Lines, Line{p.ID, qty, p.Price})
    return nil
}
// service orchestrates; handler stays a few lines
func (s *OrderService) Checkout(ctx context.Context, actor User, cart Cart) (Order, error) { /*...*/ }
```

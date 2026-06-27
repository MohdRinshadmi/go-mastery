# Phase 4 — Backend Development (Days 16–20)

`net/http`, middleware & JWT, RBAC, config & structured logging, repositories & Postgres/Redis, clean architecture. Self-quiz: answer aloud, then expand.

---

### 1. What does `http.Handler` look like, and why is it so small? What is `http.HandlerFunc`?

<details><summary>Answer</summary>

`http.Handler` is one method: `ServeHTTP(w http.ResponseWriter, r *http.Request)`. It's tiny so that *anything* — a struct, a router, a middleware wrapper — can be a handler and they all compose uniformly. `http.HandlerFunc` is an **adapter type**: `type HandlerFunc func(ResponseWriter, *Request)` with a `ServeHTTP` method that just calls the function itself. That lets a plain function satisfy the `Handler` interface — a clean demonstration of "method on a func type" turning behavior into a value.
</details>

---

### 2. What happens if you call `w.Write()` before `w.Header().Set()`?

<details><summary>Answer</summary>

The first `Write` (or `WriteHeader`) **flushes the status line and headers** to the client. Any `Header().Set(...)` after that is **silently ignored** — the headers are already on the wire. Same for setting the status code: you must `w.WriteHeader(code)` *before* the first body write, or you implicitly send `200`. Order is always: set headers → `WriteHeader(status)` → write body.
</details>

---

### 3. What are Go 1.22's new ServeMux routing features?

<details><summary>Answer</summary>

The stdlib `http.ServeMux` gained **method-based** patterns and **path wildcards**: `mux.HandleFunc("GET /products/{id}", h)`, with `r.PathValue("id")` to read the segment. It also supports `{path...}` for trailing wildcards and resolves overlapping patterns by specificity. This removes the need for a third-party router for many CRUD services — you get typed methods and path params without Gin/chi, though heavier needs (groups, rich middleware) may still warrant a framework.
</details>

---

### 4. Why always set `ReadTimeout`/`WriteTimeout` on `http.Server`? What if you don't?

<details><summary>Answer</summary>

The zero-value `http.Server` has **no timeouts**, so a slow or malicious client that opens a connection and dribbles bytes (or never finishes reading) ties up a goroutine and connection **indefinitely** — a trivial Slowloris DoS that exhausts your server. Always set `ReadTimeout`, `WriteTimeout`, `IdleTimeout` (and `ReadHeaderTimeout`) to bound how long any phase can take:

```go
srv := &http.Server{Addr: ":8080", Handler: mux,
    ReadTimeout: 5*time.Second, WriteTimeout: 10*time.Second, IdleTimeout: 60*time.Second}
```
</details>

---

### 5. What does `c.Abort()` do in Gin middleware, and how does it differ from `return`?

<details><summary>Answer</summary>

`c.Abort()` sets a flag that **stops the rest of the handler chain** from running after the current middleware returns — subsequent handlers are skipped. A bare `return` only exits the *current* middleware function; if you've already called `c.Next()` or the chain continues, the next handlers still run. So for auth rejection you must `c.AbortWithStatus(401)` (or set the response then `c.Abort()`), not just `return`, or the protected handler executes anyway. `AbortWithStatus` is `Abort` plus writing the status code.
</details>

---

### 6. When choose stdlib `net/http` over Gin (or vice versa)?

<details><summary>Answer</summary>

Use **stdlib** for small services, libraries, or when you want zero dependencies and full control — Go 1.22 routing covers most CRUD now. Use **Gin** (or chi/echo) when you want ergonomic routing groups, built-in JSON binding/validation, a middleware ecosystem, and less boilerplate on a larger API surface. The honest answer in an interview: stdlib first for simple/long-lived services, reach for a framework when the boilerplate or routing complexity actually starts hurting — not by default.
</details>

---

### 7. Explain JWT structure. Is the payload encrypted? What if I base64-decode it?

<details><summary>Answer</summary>

A JWT is three base64url parts joined by dots: **header.payload.signature**. The payload is only **base64-encoded, not encrypted** — anyone can decode and read the claims, so never put secrets (passwords, PII) in it. The **signature** (HMAC or RSA/ECDSA over header+payload) proves *integrity and authenticity*: the server verifies it with its key, so a tampered payload fails verification. JWT gives you "this token wasn't forged or altered," not confidentiality.
</details>

---

### 8. What is the `alg:none` attack, and how do you prevent it in `golang-jwt`?

<details><summary>Answer</summary>

An attacker crafts a token with header `"alg":"none"` and **no signature**, hoping the server skips verification and trusts the claims. Prevent it by **explicitly validating the signing method** in your key function and never accepting `none`:

```go
jwt.Parse(tok, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
    }
    return secret, nil
})
```
A related attack is **algorithm confusion** (passing an RSA public key where HMAC is expected) — pinning the exact method type blocks both.
</details>

---

### 9. Why short-lived JWTs + refresh tokens instead of one long-lived token?

<details><summary>Answer</summary>

JWTs are **stateless**, so they can't be revoked before they expire — a stolen long-lived token is valid until it dies. Short-lived access tokens (minutes) cap the blast radius of a leak, while a **refresh token** (long-lived, stored server-side and revocable) lets the user obtain new access tokens without re-logging-in. Revoking the refresh token cuts off renewal, so you get statelessness's performance *and* a revocation path. The tradeoff is the extra refresh round-trip and refresh-token storage.
</details>

---

### 10. How do you pass data from middleware to a handler in Gin (and stdlib)? Design RBAC for `customer`/`vendor`/`admin`.

<details><summary>Answer</summary>

In Gin, `c.Set("userID", id)` in middleware and `c.MustGet("userID")` in the handler; in stdlib, attach to the request context with `r = r.WithContext(context.WithValue(...))`. For **RBAC**, after the auth middleware sets the user's role, a `RequireRole(...)` middleware checks it before the handler runs. Map endpoints to least-privilege roles: `customer` → browse products, place/read *own* orders; `vendor` → manage *their own* products and view their sales; `admin` → manage all users/products and view everything. Crucially, enforce **ownership** (a customer reading only their own order) at the **service layer**, not just the route — role gates the door, ownership gates the row.
</details>

---

### 11. Why return proper HTTP status codes instead of `200` + error body?

<details><summary>Answer</summary>

The status code is the **machine-readable contract**: clients, proxies, load balancers, retry logic, and monitoring all key off it. `200`-with-error-in-body lies to every one of them — retries won't fire, caches misbehave, dashboards show 100% success while users see failures. Use the semantics: `400` bad input, `401` unauthenticated, `403` unauthorized, `404` not found, `409` conflict, `422` validation, `500` server fault, `503` not ready. The body explains; the code *decides*.
</details>

---

### 12. What does 12-factor say about config, and why env vars over files? Why validate at startup?

<details><summary>Answer</summary>

12-factor says **config that varies per deploy lives in the environment**, strictly separated from code. Env vars win because they're language-agnostic, easy to inject in containers/orchestrators, and keep secrets out of the repo and image (no committed config file to leak). You **validate at startup and fail fast** — parse, type-check, and bound-check every value when the process boots — so a missing `DATABASE_URL` or malformed port **crashes immediately and loudly** at deploy time, not three hours later on the first request that needs it. A process that started must be a process that can run.
</details>

---

### 13. What is structured logging and why does it beat `fmt.Println`? When `Error` vs `Warn` vs `Info`?

<details><summary>Answer</summary>

Structured logging emits **machine-parseable key/value records** (JSON via `log/slog`) instead of free-text, so you can filter, search, and aggregate by field (`user_id`, `request_id`, `status`) in your log platform. `fmt.Println` produces unqueryable prose that's useless at scale. Levels: **`Info`** = normal noteworthy events (request served, job done); **`Warn`** = recoverable anomalies / degraded-but-handled (retry succeeded, fell back to cache); **`Error`** = a failure that needs attention (request failed, dependency down). Never log secrets, passwords, tokens, or full PII.
</details>

---

### 14. How do correlation/request IDs help during an incident? Why pass `r.Context()` into service and DB calls?

<details><summary>Answer</summary>

A **correlation ID** generated per request and attached to every log line (and propagated downstream) lets you reconstruct the *entire* path of one request across handlers, services, and even other services from your logs — turning a haystack into a thread you can pull. You pass `r.Context()` down so (1) that request-scoped logger/ID and tracing flow through, and (2) **cancellation/timeout propagates**: if the client disconnects or the request times out, the DB query gets cancelled instead of running on as wasted work. Context is the spine that carries deadline, cancellation, and correlation through every layer.
</details>

---

### 15. What is the repository pattern, which layer owns the interface, and why? Why map `pgx.ErrNoRows` at the boundary?

<details><summary>Answer</summary>

The repository pattern hides persistence behind a domain-shaped interface (`UserRepository` with `GetByID`, `Save`), so the **service layer depends on the abstraction, not on SQL**. The **service/domain layer owns the interface** (dependency inversion): the concrete Postgres repo implements it, so you can swap in an in-memory repo for tests or a different DB later without touching business logic. You **map `pgx.ErrNoRows` to a domain error** (`ErrUserNotFound`) *at the repository boundary* so the leak of "we use pgx" stops there — upper layers branch on the stable domain error with `errors.Is`, and changing drivers doesn't ripple upward.
</details>

---

### 16. Why parameterized queries, why a connection pool, and what's the N+1 problem?

<details><summary>Answer</summary>

**Parameterized queries** (`$1`, `$2` placeholders) keep data separate from SQL text, which **prevents SQL injection** — the driver never interpolates user input into the statement. A **connection pool** (e.g., `pgxpool`) reuses a bounded set of connections instead of opening one per request, avoiding the high cost of TCP+auth handshakes and protecting the DB from connection exhaustion under load. **N+1** is issuing one query for a list (N rows) then one query *per row* to fetch a relation — N+1 round-trips that murder latency; fix with a `JOIN`, a single `WHERE id = ANY($1)` batch query, or a dataloader that batches the lookups.
</details>

---

### 17. Describe cache-aside. Why is a TTL essential and why is invalidation hard?

<details><summary>Answer</summary>

**Cache-aside (lazy loading):** on read, check the cache; on miss, load from the DB, populate the cache, and return. The app owns the cache; the cache doesn't know about the DB. A **TTL is essential** because it's your safety net for staleness — even if you forget to invalidate on every write path, entries self-expire, bounding how wrong the cache can be. Invalidation is hard because it's the classic "two sources of truth" problem: races between a write and a concurrent read can repopulate a stale value, and you must invalidate on *every* path that mutates the underlying data — miss one and you serve lies until the TTL saves you. ("There are only two hard problems...")
</details>

---

### 18. Clean architecture: which way do dependencies point, what may the domain import, and why must the service layer not import `net/http`?

<details><summary>Answer</summary>

Dependencies point **inward**: transport → service → domain. The **domain imports nothing** from outer layers (no `net/http`, no SQL driver, no framework) — it's pure business types and rules, so it's stable and reusable. The **service layer must not import `net/http`** because business logic shouldn't know whether it's invoked over HTTP, gRPC, a CLI, or a test — coupling it to `net/http` makes it untestable without spinning up a server and impossible to reuse behind a different transport. You test the service directly by calling its methods with plain Go values and injected fake repositories.
</details>

---

### 19. What does `internal/` enforce, where is the composition root, and how do domain errors become HTTP status codes?

<details><summary>Answer</summary>

A package under `internal/` can only be imported by code **rooted at its parent directory** — the compiler forbids external/foreign packages from importing it, so it enforces your layering and keeps implementation details private to the module. The **composition root** is `main`/`cmd/api`: the one place that constructs concrete repos, services, and handlers and wires them, centralizing all dependency decisions. **Domain errors map to status codes in exactly one place** — the transport layer's error-mapping function translates `ErrNotFound`→404, `ErrValidation`→422, `ErrUnauthorized`→403, default→500 — so the mapping is consistent and the service stays HTTP-ignorant.
</details>

---

### 20. When is this layering overkill, and when is it essential?

<details><summary>Answer</summary>

It's **overkill** for a tiny tool, a prototype, or a single-handler service — the indirection and interface ceremony cost more than they save, and you can collapse to one package. It's **essential** when multiple people work the codebase, the domain is non-trivial, you need to swap infrastructure (DB, transport) or test business logic in isolation, and the project will live for years. The senior judgment is matching ceremony to lifespan and team size — start with a **modular monolith** with clear internal boundaries and extract only when the pain is real.
</details>

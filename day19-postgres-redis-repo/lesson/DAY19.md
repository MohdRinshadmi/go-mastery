# Day 19 — PostgreSQL, the Repository Pattern, and Redis

> Mentor note: Today your service grows a memory. The key idea isn't "how to run SQL" — it's the **Repository pattern**: your business logic talks to an *interface* (`UserRepository`), never to `*sql.DB` directly. That one indirection is what lets you unit-test the service with an in-memory fake (Day 9), swap Postgres for something else, and keep SQL out of your handlers. We'll write the interface, an in-memory implementation that *runs offline*, and the real Postgres implementation as the reference you'd ship.

---

## 1. The Repository pattern — why it's the backbone of clean architecture

Layers (top to bottom): **HTTP handler → Service (business logic) → Repository (data access) → DB.** The repository is an interface defined by the layer *above* it (the service), in terms the *domain* understands:

```go
type User struct {
    ID    string
    Email string
    Name  string
}

// The service depends on THIS, not on *sql.DB:
type UserRepository interface {
    Create(ctx context.Context, u User) error
    GetByID(ctx context.Context, id string) (User, error)
    GetByEmail(ctx context.Context, email string) (User, error)
}
```

Now:
- **Service code** is pure logic, testable with a fake repo — no database in unit tests.
- **Swappable**: `PostgresUserRepo`, `InMemoryUserRepo`, `RedisUserRepo` all satisfy the interface.
- **SQL is contained** in one place, not smeared across handlers.

**Senior take:** "Accept interfaces" (Day 6/7) pays off here. The repository interface is owned by the consumer (service), not the implementer (postgres package) — that's the dependency-inversion principle. Juniors put `db.Query` in the HTTP handler; in six months that's untestable and impossible to change. Define the interface at the boundary.

## 2. database/sql + pgx

Go's `database/sql` is a generic interface; you add a *driver*. For Postgres the modern choice is **pgx** (`github.com/jackc/pgx/v5`), used via its stdlib adapter or its native pool.

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
// ...
row := pool.QueryRow(ctx, `SELECT id, email, name FROM users WHERE id=$1`, id)
var u User
if err := row.Scan(&u.ID, &u.Email, &u.Name); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return User{}, ErrUserNotFound   // map driver error -> domain error
    }
    return User{}, fmt.Errorf("get user: %w", err)
}
```

### Non-negotiable rules
1. **Always pass `ctx`** to every query — cancellation + timeouts (Day 13).
2. **Always use parameterized queries** (`$1`, `$2`) — string-concatenating user input is SQL injection. Never `fmt.Sprintf` a query with input.
3. **Map driver errors to domain errors** at the repository boundary (`pgx.ErrNoRows` → `ErrUserNotFound`) so the service doesn't know about pgx.
4. **Connection pooling**: use a pool (`pgxpool`), configure max conns; don't open a connection per request.
5. **Close rows / use `defer rows.Close()`** on multi-row queries, and check `rows.Err()` after iterating.

### Migrations
Schema changes are versioned SQL files applied in order (tools: `golang-migrate`, `goose`, `atlas`). Never hand-edit prod schemas; migrations are code, reviewed and committed. Each migration has up/down.

**Senior take:** The repository returns *domain* types and *domain* errors. If `pgx.ErrNoRows` leaks up to your HTTP handler, the abstraction failed — your handler now depends on the database driver. Translate at the boundary.

## 3. Redis — caching and more

Redis is an in-memory key-value store used for **caching, sessions, rate-limit counters, queues, locks**. Client: `github.com/redis/go-redis/v9`.

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
rdb.Set(ctx, "user:42", jsonBytes, 5*time.Minute) // value + TTL
val, err := rdb.Get(ctx, "user:42").Result()
if errors.Is(err, redis.Nil) { /* cache miss */ }
```

### Cache-aside (the default pattern) — full treatment Day 28
1. Read from cache. Hit → return.
2. Miss → read DB, write to cache with a TTL, return.
3. On write/update → invalidate (delete) the cache key.

A `CachingUserRepo` can *wrap* the Postgres repo (decorator pattern via composition) — same interface, adds a Redis layer. Clean architecture lets you bolt caching on without touching the service.

**Senior take:** Always set a TTL. A cache without expiry is a memory leak and a stale-data bug waiting to happen. And caching is a correctness/consistency decision, not just a speed one — invalidation is the hard part (Day 28).

## Common mistakes
1. SQL in HTTP handlers; no repository interface → untestable, unswappable.
2. String-built queries → SQL injection. Always parameterize.
3. No `ctx` on queries → no timeout/cancellation; a slow query hangs the request.
4. Leaking `pgx.ErrNoRows`/`sql.ErrNoRows` to upper layers instead of a domain error.
5. Opening a DB connection per request instead of a pool.
6. Redis cache with no TTL; or caching then forgetting to invalidate on writes.
7. Not checking `rows.Err()` after a row iteration.

## Performance
- Connection pool sizing matters: too small → queries queue; too large → overwhelm Postgres. Tune `MaxConns`.
- `N+1 queries` (one query per item in a loop) is the classic killer — batch with `IN ($1,$2,...)` or a join.
- Indexes on columns you filter/join by; an unindexed `WHERE email=` does a full table scan.
- Redis cache turns a 5ms DB read into a 0.2ms memory read — but only helps read-heavy keys; measure hit rate.

---

## Expert Thinking Mode — "store the data"

- **Beginner:** "`db.Query` right in the handler."
- **Senior:** "Repository interface owned by the service; Postgres impl maps driver errors to domain errors; ctx on every query; parameterized SQL; pool. In-memory fake for tests."
- **Staff:** "Read/write split? Cache-aside with TTL and invalidation? Migration strategy and backward-compatible schema changes for zero-downtime deploys? N+1 and index review."
- **Architect:** "Data store choice (SQL vs KV vs document), consistency model, sharding, and the read/write path are system-defining. The repository boundary is what keeps those decisions from leaking into business logic."

---

## Real-world use

- **Every Go backend** uses the repository pattern (or a close cousin) to keep persistence swappable and testable.
- **pgx** is the de-facto Postgres driver in Go; **go-redis** the standard Redis client.
- **Cache-aside with Redis** in front of Postgres is the single most common read-scaling pattern in web backends.
- **Migrations** (`golang-migrate`/`goose`) gate schema changes in CI/CD (Phase 5).

---

## Interview Questions

1. What is the repository pattern and which layer owns the interface? Why (dependency inversion)?
2. Why parameterized queries? What attack do they prevent?
3. Why pass `ctx` to every DB call?
4. Why map `pgx.ErrNoRows` to a domain error at the repository boundary?
5. Describe cache-aside. Why is a TTL essential and why is invalidation hard?
6. What is an N+1 query problem and how do you fix it?
7. Why a connection pool instead of a connection per request?

---

## Your tasks

`../exercises/` defines a `UserRepository` interface and asks you to: (1) implement `InMemoryUserRepo` (runs offline, used in tests), (2) sketch the `PostgresUserRepo` method signatures mapping `pgx.ErrNoRows`→`ErrUserNotFound`, and (3) a challenge: a `CachingUserRepo` that wraps another `UserRepository` with a tiny in-memory TTL cache (same interface — composition!). A `docker-compose.yml` for Postgres+Redis is provided for when you wire the real thing. The runnable demo uses the in-memory repo so it works without any database. Reference in `../solutions/`.

## Day 19 companion files

- [Debugging challenge](../debugging/README.md) — the repository that leaks its internal pointer (aliasing corruption).
- [Pitfalls](../PITFALLS.md) — Trap → Why → Fix: injection, missing ctx, leaking ErrNoRows, aliasing, per-request conns, TTL/invalidation, rows.Err(), N+1.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.

# Day 19 — Postgres, Repository & Redis Interview Questions

Model answers fold; try to answer before expanding.

---

### 1. What is the repository pattern, which layer owns the interface, and why (dependency inversion)?

<details>
<summary>Answer</summary>

The repository pattern puts all data access behind an interface that speaks in
*domain* terms (`UserRepository` with `GetByID`, `Create`, ...), so the business
logic never touches `*sql.DB`, pgx, or SQL directly. Layers run HTTP handler →
service → repository → DB.

The interface is owned by the **consumer** — the service layer — not by the
Postgres package that implements it. That's the dependency-inversion principle:
both the high-level service and the low-level Postgres code depend on an
abstraction (the interface) that the high-level code defines. The arrow of
dependency points *toward* the business logic, not toward the database.

Payoff: the service is unit-testable with an in-memory fake (no DB in tests),
implementations are swappable (`PostgresUserRepo`, `InMemoryUserRepo`,
`CachingUserRepo`), and SQL lives in one place instead of smeared across
handlers.

</details>

---

### 2. Why parameterized queries? What attack do they prevent?

<details>
<summary>Answer</summary>

They prevent **SQL injection**. When you build a query by concatenating user
input into the SQL string, the input becomes part of the executable SQL — an
attacker can supply `x' OR '1'='1` to bypass a filter or `'; DROP TABLE
users;--` to run arbitrary statements.

With parameters (`$1`, `$2`), the driver sends the SQL text and the argument
*values* to Postgres separately. The values are bound after parsing, so they can
never be interpreted as SQL syntax — only as data. `fmt.Sprintf` into a query is
the same vulnerability; never do it. Parameterization also lets the server
reuse the prepared plan.

</details>

---

### 3. Why pass `ctx` to every DB call?

<details>
<summary>Answer</summary>

`ctx` carries cancellation and deadlines. If the HTTP client disconnects or the
request's timeout fires, the context is cancelled and the in-flight query is
aborted, releasing the pooled connection. Without `ctx`, a slow query keeps
running after nobody's waiting for it, holding connections; under load these
pile up and exhaust the pool — one slow query becomes an outage. `ctx` also
carries request-scoped values (trace IDs) for observability. It belongs as the
first parameter of every repository method and is threaded straight into the
query call.

</details>

---

### 4. Why map `pgx.ErrNoRows` to a domain error at the repository boundary?

<details>
<summary>Answer</summary>

To keep the abstraction intact. If `pgx.ErrNoRows` (or `sql.ErrNoRows`) leaks up,
the service and handler must import and check a *driver* error — they now depend
on Postgres specifics, and swapping the store breaks every caller. The
repository's job is to translate driver concepts into domain concepts: a missing
row becomes `ErrUserNotFound`, a domain sentinel callers check with `errors.Is`.
The handler then maps `ErrUserNotFound` → HTTP 404 without ever knowing pgx
exists. Translate at the boundary; wrap genuinely unexpected errors with `%w`.

</details>

---

### 5. Describe cache-aside. Why is a TTL essential and why is invalidation hard?

<details>
<summary>Answer</summary>

Cache-aside (lazy loading):
1. Read from the cache. On a hit, return it.
2. On a miss, read the DB, write the value into the cache with a **TTL**, return.
3. On a write/update, **delete** (invalidate) the cache key.

A TTL is essential because it bounds both *memory* (entries expire instead of
accumulating forever — no leak) and *staleness* (even if you miss an
invalidation, the entry self-heals when it expires). It's the safety net.

Invalidation is hard because it's a distributed-consistency problem: the same
data may be cached under several keys, multiple writers race to update DB and
cache, and a crash between "write DB" and "delete cache key" leaves stale data.
There's a classic race where a concurrent read repopulates the cache with the
old value right after a delete. TTL caps how long any such bug can bite.

</details>

---

### 6. What is the N+1 query problem and how do you fix it?

<details>
<summary>Answer</summary>

You run 1 query to fetch a list of N parents, then 1 query per parent to fetch
its children — N+1 round trips total. Each round trip pays network latency, so
the endpoint scales linearly with N and crawls (501 queries for 500 items).

Fix: fetch the children in a **single batched query** using
`WHERE id = ANY($1)` / `IN ($1,$2,...)` or a `JOIN`, then group the rows in
memory. One round trip instead of N. (ORMs call the eager version
"preloading"/"eager loading".) Watch for it in any loop that issues a query.

</details>

---

### 7. Why a connection pool instead of a connection per request?

<details>
<summary>Answer</summary>

Opening a connection means a TCP handshake, TLS, and Postgres authentication —
expensive, often costing more than the query itself. Doing it per request adds
latency and, worse, Postgres has a hard `max_connections` limit; under load you
exhaust it and the database refuses new connections. A pool (e.g. `pgxpool`)
opens a bounded set of connections once at startup and lends them out, reusing
each across many requests. You create one pool, share it, and tune `MaxConns`.

</details>

---

### 8. How does `CachingUserRepo` compose with the Postgres repo (decorator/wrapping)?

<details>
<summary>Answer</summary>

`CachingUserRepo` is a **decorator**: it *implements* `UserRepository` and also
*holds* a `UserRepository` (the inner Postgres repo) plus a Redis client. Each
method does the cache-aside dance and delegates misses/writes to the inner repo:

```go
type CachingUserRepo struct {
    inner UserRepository
    rdb   *redis.Client
    ttl   time.Duration
}

func (c *CachingUserRepo) GetByID(ctx context.Context, id string) (User, error) {
    if u, ok := c.fromCache(ctx, id); ok { return u, nil }
    u, err := c.inner.GetByID(ctx, id) // delegate on miss
    if err == nil { c.toCache(ctx, u) }
    return u, err
}
```

Because it satisfies the same interface, the service is unchanged — you wrap
`NewCachingUserRepo(postgresRepo, rdb)` at composition time. This is the power of
accepting interfaces: you bolt caching on without touching business logic, and
you can stack decorators (logging, metrics, retry) the same way.

</details>

---

### 9. Why is `sql.ErrNoRows` a sentinel, and how does `errors.Is` work?

<details>
<summary>Answer</summary>

A *sentinel* error is a single exported package-level value (`var ErrNoRows =
errors.New(...)`) that callers compare against to recognize a specific condition.
It works because `errors.New` returns a pointer to a unique struct — identity
comparison is what makes it recognizable.

`errors.Is(err, target)` walks the error's `Unwrap` chain, comparing each link
to `target` (and honoring any `Is(error) bool` method). So even when a sentinel
is wrapped with `fmt.Errorf("...: %w", ErrNoRows)`, `errors.Is(err, ErrNoRows)`
still returns true. That's why you compare with `errors.Is`, not `==`: `==` only
matches an *unwrapped* sentinel and breaks the moment someone wraps it. You map
`pgx.ErrNoRows` → your own `ErrUserNotFound` sentinel so callers get the same
ergonomics on a domain error.

</details>

---

### 10. What are the tradeoffs in connection pool sizing?

<details>
<summary>Answer</summary>

Too **small**: queries queue waiting for a free connection, latency spikes, and
you under-use both app and DB even though resources are idle. Too **large**: you
overwhelm Postgres — each connection costs server memory and a backend process,
and past a point more connections *reduce* throughput due to lock and scheduler
contention (Postgres isn't great at thousands of connections).

The pool max should be sized against the DB's capacity, not the app's
concurrency, and summed across *all* app instances (10 pods × 20 conns = 200
connections hitting one DB). Common practice: a modest pool per instance, and a
server-side pooler like PgBouncer in front when many instances multiply the
count. Measure wait time and DB CPU; tune `MaxConns` from data.

</details>

---

### 11. Why use an in-memory fake repo for unit tests instead of a real database?

<details>
<summary>Answer</summary>

Because the repository interface lets you substitute a fast, deterministic
in-memory implementation, so service unit tests run with no DB: no Docker, no
network, milliseconds not seconds, no shared-state flakiness, and easy setup of
exact scenarios (including forcing `ErrUserNotFound` or an error path). You're
testing *business logic*, and the interface is the seam that makes that possible
without a database.

The tradeoff: a fake can drift from real Postgres behavior (it won't catch a SQL
typo, a constraint violation, or a transaction-isolation bug). So you complement
unit tests with a smaller set of **integration tests** that run the real
`PostgresUserRepo` against a real Postgres (e.g. testcontainers) in CI. Fakes for
breadth and speed; integration tests for fidelity at the data layer.

</details>

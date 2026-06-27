# Day 19 — Postgres, Repository & Redis Cheatsheet

Quick reference. The repository boundary is the whole game: domain types in,
domain errors out, SQL hidden inside.

---

## The interface + domain error (owned by the service layer)

```go
type User struct {
    ID    string
    Email string
    Name  string
}

// The consumer (service) defines this; implementations satisfy it.
type UserRepository interface {
    Create(ctx context.Context, u User) error
    GetByID(ctx context.Context, id string) (User, error)
    GetByEmail(ctx context.Context, email string) (User, error)
}

// Domain sentinel — what callers check, never the driver error.
var ErrUserNotFound = errors.New("user not found")
```

---

## pgx: QueryRow + Scan + map ErrNoRows

```go
func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (User, error) {
    row := r.pool.QueryRow(ctx,
        `SELECT id, email, name FROM users WHERE id = $1`, id) // parameterized
    var u User
    if err := row.Scan(&u.ID, &u.Email, &u.Name); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return User{}, ErrUserNotFound      // map at the boundary
        }
        return User{}, fmt.Errorf("get user %s: %w", id, err)
    }
    return u, nil
}
```

Multi-row: `defer rows.Close()`, loop `rows.Next()` + `rows.Scan`, then
**check `rows.Err()`**.

---

## The 5 non-negotiable rules

1. **Always pass `ctx`** to every query (cancellation + timeouts).
2. **Always parameterize** (`$1`, `$2`) — never concatenate input into SQL.
3. **Map driver errors to domain errors** at the boundary (`pgx.ErrNoRows` →
   `ErrUserNotFound`).
4. **Use a connection pool** (`pgxpool`), tune `MaxConns` — not one conn per request.
5. **`defer rows.Close()` and check `rows.Err()`** on multi-row queries.

(Bonus repo rule: return **copies**, not internal pointers/slices — no aliasing.)

---

## Parameterized query (the one habit that matters)

```go
// NO — SQL injection:
q := "SELECT ... WHERE email = '" + email + "'"

// YES — driver binds the value, never as SQL:
row := pool.QueryRow(ctx, `SELECT ... WHERE email = $1`, email)
```

---

## Cache-aside steps

1. Read cache. **Hit** → return.
2. **Miss** → read DB → write to cache **with a TTL** → return.
3. On **write/update** → **delete** (invalidate) the cache key.

```go
val, err := rdb.Get(ctx, "user:"+id).Result()
if errors.Is(err, redis.Nil) { /* miss → load DB, then Set with TTL */ }
rdb.Set(ctx, "user:"+id, jsonBytes, 5*time.Minute) // TTL is mandatory
rdb.Del(ctx, "user:"+id)                            // on update
```

---

## CachingUserRepo wrapping shape (decorator via composition)

```go
type CachingUserRepo struct {
    inner UserRepository  // wrapped repo (e.g. Postgres)
    rdb   *redis.Client
    ttl   time.Duration
}

// Same interface → service code is unchanged; just wrap at startup.
func (c *CachingUserRepo) GetByID(ctx context.Context, id string) (User, error) {
    if u, ok := c.fromCache(ctx, id); ok {
        return u, nil
    }
    u, err := c.inner.GetByID(ctx, id) // delegate on miss
    if err == nil {
        c.toCache(ctx, u) // Set with c.ttl
    }
    return u, err
}

repo := NewCachingUserRepo(postgresRepo, rdb, 5*time.Minute)
```

---

## Key terms

- **Repository pattern** — data access behind a domain-typed interface; SQL lives
  in one place, the service depends on the interface.
- **Dependency inversion** — the consumer (service) owns the interface; the
  low-level DB code depends on it, not the reverse.
- **Sentinel error** — a unique package-level `error` value (`ErrUserNotFound`)
  callers recognize with `errors.Is`.
- **Parameterized query** — `$1`/`$2` placeholders bound by the driver; input is
  data, never SQL — prevents injection.
- **Connection pool** — a bounded, reused set of DB connections opened once
  (`pgxpool`); tune `MaxConns`.
- **Cache-aside** — read cache → on miss read DB and populate cache → invalidate
  on write.
- **TTL** — time-to-live on a cache entry; bounds memory and staleness.
- **Decorator / wrapping** — a type that implements an interface and also holds
  one, adding behavior (caching, logging) via composition.
- **N+1** — 1 list query + 1 query per element; fix by batching with
  `IN ($1,...)` / `ANY($1)` or a JOIN.

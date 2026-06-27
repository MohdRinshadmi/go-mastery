# Day 19 — Postgres, Repository & Redis Pitfalls (Trap → Why → Fix)

Concrete traps you'll hit wiring a data layer in Go. Each is **Trap → Why → Fix**.

---

## 1. String-built queries → SQL injection

**Trap.**

```go
q := "SELECT id, email, name FROM users WHERE email = '" + email + "'"
row := pool.QueryRow(ctx, q)
```

**Why.** User input becomes *part of the SQL text*. An email of
`x' OR '1'='1` (or `x'; DROP TABLE users;--`) changes the query's meaning. This
is the most common, most damaging web vulnerability — and `fmt.Sprintf` into a
query is the exact same bug.

**Fix.** Pass input as **parameters**; the driver sends them separately from the
SQL, so they can never be interpreted as code.

```go
row := pool.QueryRow(ctx,
    `SELECT id, email, name FROM users WHERE email = $1`, email)
```

---

## 2. No `ctx` on queries

**Trap.**

```go
func (r *PostgresUserRepo) GetByID(id string) (User, error) {
    row := r.pool.QueryRow(context.Background(), `SELECT ... WHERE id=$1`, id)
    ...
}
```

**Why.** A query with no caller context can't be cancelled or time-limited. When
the HTTP client disconnects or the request deadline (Day 13) fires, the query
keeps running, holding a pooled connection — slow queries pile up into an outage.

**Fix.** Thread the request's `ctx` through every call. The repository method
*takes* `ctx context.Context` as its first parameter and passes it down.

```go
func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (User, error) {
    row := r.pool.QueryRow(ctx, `SELECT ... WHERE id=$1`, id)
    ...
}
```

---

## 3. Leaking `pgx.ErrNoRows` / `sql.ErrNoRows` to upper layers

**Trap.**

```go
// service or handler:
u, err := repo.GetByID(ctx, id)
if errors.Is(err, pgx.ErrNoRows) { // handler now imports pgx!
    http.Error(w, "not found", 404)
}
```

**Why.** The whole point of the repository is to hide the driver. If the handler
checks `pgx.ErrNoRows`, it depends on Postgres specifics — swap to another store
and every caller breaks. The abstraction has failed.

**Fix.** Map the driver error to a **domain sentinel** at the repository
boundary; callers check the domain error with `errors.Is`.

```go
// repository:
if err := row.Scan(&u.ID, &u.Email, &u.Name); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return User{}, ErrUserNotFound // domain error
    }
    return User{}, fmt.Errorf("get user: %w", err)
}
```

---

## 4. Returning a shared internal pointer or slice (aliasing)

**Trap.**

```go
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    return r.users[id], nil // the SAME pointer the map holds
}
```

**Why.** The caller and the repository now share one `User`. A caller mutating
the returned value (`u.Name = ...`) silently corrupts the repo's stored data.
Same trap with `[]T`: returning an internal slice lets a caller `append`/index
into your storage. This is the Day 19 debugging challenge.

**Fix.** Return a **copy** at the boundary — a value, a freshly-allocated
pointer, or a cloned slice.

```go
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    u, ok := r.users[id]
    if !ok {
        return nil, ErrUserNotFound
    }
    cp := *u
    return &cp, nil
}
```

---

## 5. A connection per request instead of a pool

**Trap.**

```go
func handler(w http.ResponseWriter, req *http.Request) {
    conn, _ := pgx.Connect(req.Context(), os.Getenv("DATABASE_URL"))
    defer conn.Close(req.Context())
    ...
}
```

**Why.** Opening a TCP connection + TLS + Postgres auth per request is slow and
exhausts the server's connection limit under load — Postgres falls over. Connect
setup can dwarf the query itself.

**Fix.** Create **one pool** at startup, share it, let it hand out and reuse
connections. Tune `MaxConns`.

```go
pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL")) // once, at startup
// pass *pgxpool.Pool into the repository; every method uses it
```

---

## 6. Redis cache with no TTL, or forgetting to invalidate on write

**Trap.**

```go
rdb.Set(ctx, "user:"+id, jsonBytes, 0) // 0 = never expires
// ...and the Update path never deletes the key.
```

**Why.** No TTL means entries live forever — an unbounded memory leak — and the
cache serves *stale* data after the row changes. Caching is a correctness
decision, not just a speed one; invalidation is the hard part.

**Fix.** Always set a TTL, and **delete the key on write/update** (cache-aside).

```go
rdb.Set(ctx, "user:"+id, jsonBytes, 5*time.Minute) // bounded
// on update:
_ = inner.Update(ctx, u)
rdb.Del(ctx, "user:"+u.ID) // invalidate; next read repopulates from DB
```

---

## 7. Not checking `rows.Err()` after iterating

**Trap.**

```go
rows, _ := pool.Query(ctx, `SELECT id, name FROM users`)
defer rows.Close()
for rows.Next() {
    rows.Scan(&u.ID, &u.Name)
    out = append(out, u)
}
return out, nil // never checked rows.Err()
```

**Why.** `rows.Next()` returns `false` both at the *normal end* and when
iteration *failed mid-stream* (connection dropped, decode error). Skipping
`rows.Err()` means you silently return a truncated result set as if it were
complete.

**Fix.** Always check `rows.Err()` after the loop, and `defer rows.Close()`.

```go
for rows.Next() {
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return nil, err
    }
    out = append(out, u)
}
if err := rows.Err(); err != nil { // catches mid-iteration failures
    return nil, err
}
return out, nil
```

---

## 8. N+1 queries

**Trap.**

```go
orders, _ := repo.OrdersFor(ctx, userID)
for _, o := range orders {
    item, _ := repo.GetItem(ctx, o.ItemID) // one query PER order
    ...
}
```

**Why.** One query for the list plus one per element = N+1 round trips. With 500
orders that's 501 queries; network latency dominates and the endpoint crawls.

**Fix.** Batch the children in a single query with `IN ($1,$2,...)` (or a JOIN),
then group in memory.

```go
rows, _ := pool.Query(ctx,
    `SELECT id, name FROM items WHERE id = ANY($1)`, itemIDs) // one round trip
```

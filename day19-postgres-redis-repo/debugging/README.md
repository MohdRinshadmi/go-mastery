# Debugging Challenge — The Repository That Leaks Its Internals

An in-memory `UserRepository` looks correct: `GetByID` finds the user, returns
it, and returns a domain sentinel when the id is missing. Yet after one caller
touches a returned user, *every later read of that id is wrong* — and nothing
errors. This is the aliasing gotcha of Day 19.

## Symptom

Fetch user `u1` (`"Alice"`), do an unrelated mutation on the returned value, then
fetch `u1` again — and the *second, independent* read comes back `"HACKED"`. The
repository's own stored data was silently corrupted by a caller it never trusted
with write access.

## Repro

Bugged (wrong output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day19-postgres-redis-repo/debugging/bugged
go run .
```

Expected (buggy) output:

```
fetch #1: u1 -> "Alice"
fetch #2: u1 -> "HACKED"
CORRUPTED: caller's mutation leaked into the repository's stored data
```

Fixed (correct output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day19-postgres-redis-repo/debugging/fixed
go run .
```

Expected (correct) output:

```
fetch #1: u1 -> "Alice"
fetch #2: u1 -> "Alice"
OK: stored data is intact
missing id -> ErrUserNotFound (mapped at the boundary)
```

## Hint

The repo stores `map[string]*User` and `GetByID` returns `r.users[id]`. What
does the caller get — a *copy* of the user, or the *same pointer* the map holds?
When two variables hold the same pointer, whose `User` does `u.Name = "HACKED"`
modify?

<details>
<summary>Solution & why</summary>

The repository hands out a pointer that aliases its own storage. The map value
*is* a `*User`; returning it gives the caller the exact pointer the repository
keeps. There is now only one `User` in memory, reachable from two places — the
map and the caller. The caller's `u1.Name = "HACKED"` writes through that shared
pointer, so the repository's stored user mutates too. The next `GetByID` returns
the same corrupted pointer.

```go
// BUG: returns the stored pointer — caller and repo share one User
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    u, ok := r.users[id]
    if !ok {
        return nil, ErrUserNotFound
    }
    return u, nil // aliasing: same *User the map points at
}
```

The fix is to return a *copy*. Dereference the stored pointer to copy the struct
value, then return the address of that fresh copy. The caller's `*User` is its
own object; mutating it cannot reach the repository's storage.

```go
// FIX: copy the value, return a pointer to the copy
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    u, ok := r.users[id]
    if !ok {
        return nil, ErrUserNotFound
    }
    cp := *u
    return &cp, nil
}
```

Why this matters beyond a toy fake: a real `PostgresUserRepo` is naturally safe
here — `row.Scan` writes into a *fresh* local `User` each call, so callers never
share storage. The bug is easy to introduce the moment you add an in-memory
cache or fake that hands back internal pointers/slices. The same trap bites
slices: returning `r.internalSlice` lets a caller `append` or index-assign into
your storage. Return copies (or freshly allocated slices) at the boundary.

**Rules of thumb:**

- A repository returns **values or freshly-allocated copies**, never a pointer or
  slice that aliases its internal storage. The boundary owns its data.
- If a method returns `*T` or `[]T` sourced from a field, ask "can the caller
  mutate this and reach back into me?" If yes, copy before returning.
- `go vet` will *not* catch this — there's no type error, just shared state.
  Treat "returning an internal pointer/slice" as a code-review smell.
- (Companion rule, also modeled here: map the driver's "no rows" into a domain
  sentinel — `ErrUserNotFound` — so callers use `errors.Is` and never learn
  about `pgx.ErrNoRows`.)

</details>

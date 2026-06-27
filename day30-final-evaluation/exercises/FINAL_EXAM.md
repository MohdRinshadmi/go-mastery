# Final Exam — 30-Day Go Mastery

Do this closed-book first, then check `../solutions/ANSWER_KEY.md`. Score with
the rubric at the bottom. Bring your answers + code for a real assessment.

---

## Section A — Theory (2 pts each, 30 pts)

1. What is the zero value of a `string`, a `slice`, and a `map`? Which panics on write?
2. After `b := a[:2]; b = append(b, 99)`, why might `a` change? How do you prevent it?
3. Why are errors returned values in Go instead of exceptions? Name one cost.
4. `%w` vs `%v` in `fmt.Errorf` — when do you choose each?
5. Why does Go prefer composition over inheritance? How do interfaces enable it?
6. When are generics the right tool, and when do they hurt readability?
7. What is a data race? Why can't you find it reliably by reading code?
8. Channel vs mutex — when each?
9. Who closes a channel, and what are the two ways to misuse `close`?
10. What does `context.Context` propagate, and what are its usage rules?
11. Liveness vs readiness probes — what does each control, and the danger of conflating them?
12. State the CAP theorem. Why is there no "CA" system in practice?
13. At-least-once delivery forces what property in consumers, and why?
14. Describe cache-aside. Why is a TTL essential even with write-invalidation?
15. When do you choose synchronous gRPC vs asynchronous events between services?

## Section B — Coding (10 pts each, 30 pts)

**B1.** Write a generic `GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T`.
Include a table-driven test.

**B2.** Implement a concurrency-safe LRU-ish cache: `Set(k,v)`, `Get(k)(v,ok)` with a
max size; on overflow evict the oldest-inserted key. Must pass `go test -race`.

**B3.** Write a bounded worker pool that, given `[]string` URLs and a `check func(ctx,string) error`,
returns `map[string]error`, with per-check timeout via context and concurrency limited to N.

## Section C — Debugging (10 pts each, 20 pts)

**C1.** This leaks a goroutine. Explain why and fix it:
```go
func first(urls []string) string {
    ch := make(chan string)
    for _, u := range urls {
        go func(u string) { ch <- fetch(u) }(u)
    }
    return <-ch // returns the first; the other goroutines block forever on send
}
```

**C2.** This races and sometimes prints < 1000. Explain and fix two ways (mutex and atomic):
```go
counter := 0
var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() { defer wg.Done(); counter++ }()
}
wg.Wait()
fmt.Println(counter)
```

## Section D — Code Review (20 pts)

Review this handler as if it were a PR. List every issue (correctness, security,
idioms, architecture) you'd raise:
```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    row := db.QueryRow("SELECT name FROM users WHERE id = " + id)
    var name string
    row.Scan(&name)
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(map[string]string{"name": name})
}
```
(Hint: there are at least 6 issues spanning SQL injection, error handling,
status codes, context, and layering.)

---

## Rubric (100 pts)
- 90–100: Senior-track. You reason about trade-offs and write idiomatic, safe, tested Go.
- 75–89: Job-ready (mid). Solid fundamentals + concurrency + architecture; minor gaps.
- 60–74: Advanced beginner. Correct Go; needs more depth on concurrency/distributed/perf.
- < 60: Revisit the phases where you lost the most points; redo those exercises.

After scoring, write down your **three weakest areas** — that's your next study plan.

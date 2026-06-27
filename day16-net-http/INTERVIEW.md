# Day 16 — net/http Interview Questions

Answer out loud before expanding. Staff-level signal isn't reciting the API —
it's knowing *why* the stdlib is shaped this way and what it costs at runtime.

---

### 1. What does the `http.Handler` interface look like, and why is it so small?

<details>
<summary>Answer</summary>

```go
type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}
```

One method. That smallness is the whole point: anything that knows how to turn
a `*Request` into bytes on a `ResponseWriter` is a handler. A function, a
struct with state, a closure over a database pool — all become handlers. It
makes **middleware** trivial (a function `func(Handler) Handler`), it makes
routers swappable, and it lets unrelated libraries interoperate without a shared
framework. Go's philosophy: small interfaces compose into large systems. A
big abstract base class (Java's `HttpServlet`) couples you to an inheritance
hierarchy; a one-method interface couples you to nothing.

</details>

---

### 2. What is `http.HandlerFunc` and what pattern does it demonstrate?

<details>
<summary>Answer</summary>

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

It's a function *type* with a `ServeHTTP` method that just calls itself. This is
the **adapter pattern**: it adapts a plain function to satisfy the `Handler`
interface so you can write ordinary functions and still use them anywhere a
`Handler` is required. `http.HandleFunc(pattern, fn)` does the conversion
`http.HandlerFunc(fn)` for you. It's also a neat demonstration that in Go,
methods can be defined on any named type — not just structs.

</details>

---

### 3. What happens if you call `w.Write()` before `w.Header().Set()`?

<details>
<summary>Answer</summary>

The header set is lost. The first `Write` implicitly calls
`WriteHeader(http.StatusOK)`, which sends the status line and **freezes** the
header map. Any header you set after that point never goes out, and any later
`WriteHeader` is ignored with a `superfluous response.WriteHeader call` log.
The contract is strict and ordered: `Header().Set(...)` → `WriteHeader(status)`
→ `Write(body)`. Because the response is a stream, once body bytes are committed
you cannot retroactively change what came before them.

</details>

---

### 4. What are the new routing features in Go 1.22's ServeMux?

<details>
<summary>Answer</summary>

Go 1.22 taught the stdlib `ServeMux` two things it couldn't do before:

- **Method matching** in the pattern: `mux.HandleFunc("GET /products", ...)` —
  no more `if r.Method != "GET"` boilerplate, and a mismatched method returns
  `405 Method Not Allowed` automatically.
- **Wildcard path segments**: `GET /products/{id}` captures a segment, read in
  the handler with `r.PathValue("id")`. A trailing `{rest...}` matches the
  remainder of the path.

It also defined **precedence** (more specific patterns win, regardless of
registration order) and `405`/`Allow`-header handling. This closed most of the
gap that pushed people to third-party routers for simple services.

</details>

---

### 5. Why should you always set `ReadTimeout` and `WriteTimeout` on `http.Server`?

<details>
<summary>Answer</summary>

Because the defaults are zero, which means *no timeout*. Without them a slow or
malicious client can hold a connection open indefinitely — a Slowloris attack
trickling one byte at a time — consuming a goroutine and a file descriptor per
connection until you exhaust them. That's a denial-of-service vector.
`ReadTimeout` bounds how long reading the entire request (including body) may
take; `WriteTimeout` bounds sending the response; `IdleTimeout` bounds keep-alive
idle time. `http.ListenAndServe(addr, nil)` has none of these — fine for a
tutorial, an instant review flag in production.

</details>

---

### 6. What does `c.Abort()` do in Gin middleware, and how does it differ from just returning?

<details>
<summary>Answer</summary>

In Gin, middleware and handlers run as a **chain**. `return` only exits the
current function; the next handler in the chain still runs. `c.Abort()` sets a
flag so that the *remaining* handlers in the chain are **not** invoked — it stops
the chain. The common idiom in auth middleware is:

```go
if !authorized {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "nope"})
    c.Abort() // or c.AbortWithStatusJSON(...)
    return
}
c.Next()
```

`return` after `c.Abort()` exits the current middleware; `c.Abort()` ensures
nothing downstream runs. Forgetting `c.Abort()` and only `return`ing is a real
auth-bypass bug — the request continues to the protected handler.

</details>

---

### 7. When would you choose stdlib `net/http` over Gin?

<details>
<summary>Answer</summary>

Reach for stdlib when the service is small and the dependency cost outweighs the
ergonomics: an internal microservice with a handful of endpoints, a service
called only by known clients, or anything where you want zero external
dependencies and full control of the `ResponseWriter`. Since Go 1.22's routing
upgrade, method+wildcard matching covers most simple REST needs natively. Reach
for Gin (or Chi) when you have 20+ endpoints, want route groups with shared
middleware, need JSON binding/validation, or want automatic panic recovery —
i.e., when the boilerplate the framework removes is real and recurring. "Learn
stdlib first" is also a legitimate reason: you can't debug a framework you don't
understand.

</details>

---

### 8. How does `httptest.NewRecorder` work, and why is it the right way to test handlers?

<details>
<summary>Answer</summary>

`httptest.NewRecorder()` returns a `*httptest.ResponseRecorder`, an in-memory
implementation of `http.ResponseWriter`. Instead of writing to a socket it
records into struct fields: `Code` (status), `HeaderMap` (headers), and `Body`
(a `*bytes.Buffer`). You call your handler directly —
`handler.ServeHTTP(rec, req)` — with a request from
`httptest.NewRequest(...)`, then assert on `rec.Code`, `rec.Body.String()`,
`rec.Header()`. No port, no network, fully deterministic and fast. That's why
it's how you catch the "lost status code" bug in CI: `rec.Code` exposes the real
committed status. For an end-to-end test through a real server, use
`httptest.NewServer` instead, which spins up a server on a random local port.

</details>

---

### 9. What does each incoming request cost the server in terms of goroutines and memory?

<details>
<summary>Answer</summary>

Go's `net/http` server runs **one goroutine per connection** (and the handler
runs on that goroutine). This is what gives you "concurrency for free" — you
write synchronous-looking handler code and the runtime multiplexes thousands of
them. The cost: each goroutine starts with roughly an 8 KB stack (growable). At
100K concurrent connections that's ~800 MB just for stacks, before any
per-request allocations, buffers, or the read/write buffers the server keeps per
connection. For typical APIs this is a non-issue; for extreme connection counts
(hundreds of thousands idle), the per-goroutine overhead is why people look at
event-loop designs like `fasthttp`. The practical implication: bound concurrency
and set timeouts so you don't accumulate stuck goroutines.

</details>

---

### 10. How do you shut a Go HTTP server down gracefully?

<details>
<summary>Answer</summary>

Use `srv.Shutdown(ctx)`, not just killing the process. `Shutdown` stops
accepting new connections, then waits for in-flight requests to finish (up to
the context's deadline) before returning. The idiom:

```go
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// wait for SIGINT/SIGTERM
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Printf("graceful shutdown failed: %v", err)
}
```

Note `ListenAndServe` returns `http.ErrServerClosed` on a clean shutdown — that
specific error is expected, not a failure. `Shutdown` is what lets you do
zero-downtime deploys: drain existing requests instead of dropping them.

</details>

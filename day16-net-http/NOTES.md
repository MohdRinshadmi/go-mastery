# Day 16 — net/http Cheatsheet

Quick reference for the stdlib HTTP server. Keep it open while you build the
exercises.

---

## The `http.Handler` interface

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

Anything with this one method is a handler: a struct (stateful), a closure, a
function (via `HandlerFunc`). Middleware is `func(http.Handler) http.Handler`.

## The `http.HandlerFunc` adapter

```go
type HandlerFunc func(http.ResponseWriter, *http.Request)
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) { f(w, r) }

func index(w http.ResponseWriter, r *http.Request) { /* ... */ }

mux.Handle("GET /", http.HandlerFunc(index)) // explicit adapter
mux.HandleFunc("GET /", index)               // does the adapter for you
```

## ServeMux — Go 1.22 patterns

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /products", listProducts)
mux.HandleFunc("POST /products", createProduct)
mux.HandleFunc("GET /products/{id}", getProduct)

func getProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id") // Go 1.22+
    // ...
}
```

| Pattern | Matches |
|---------|---------|
| `/` | All paths (catch-all) |
| `/api/` | `/api/` and everything under it (trailing slash = subtree) |
| `/api/users` | Exact path only |
| `GET /api/users` | GET requests to that exact path (else 405) |
| `GET /api/users/{id}` | GET with one wildcard segment (`r.PathValue("id")`) |
| `GET /files/{path...}` | GET; `{path...}` captures the rest of the path |

Precedence: the **most specific** matching pattern wins, regardless of
registration order. Method mismatch returns `405` with an `Allow` header
automatically.

## `http.Server` with timeouts

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,   // read entire request incl. body
    WriteTimeout: 10 * time.Second,  // send the response
    IdleTimeout:  120 * time.Second, // keep-alive idle
}
log.Fatal(srv.ListenAndServe())
```

Never ship `http.ListenAndServe(addr, nil)` to production — zero timeouts = DoS
vector, and `nil` uses global `DefaultServeMux`.

Graceful shutdown:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx) // drains in-flight requests; ListenAndServe returns http.ErrServerClosed
```

## Response-writing helpers

```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json") // 1. headers
    w.WriteHeader(status)                              // 2. status
    json.NewEncoder(w).Encode(data)                    // 3. body
}

// Errors: http.Error sets Content-Type, writes status + body in order.
http.Error(w, "not found", http.StatusNotFound)
return // ALWAYS return after http.Error
```

## Reading requests safely

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // cap at 1 MiB
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}
```

## The ordering rule (burn this in)

```
Header().Set(...)  →  WriteHeader(status)  →  Write(body)
```

The first `Write` implicitly sends `200 OK` and freezes the headers. Want any
other status or header? Do it before the first body byte.

## Testing without a port

```go
rec := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodGet, "/products/42", nil)
req.SetPathValue("id", "42") // when calling a handler directly
handler.ServeHTTP(rec, req)
// assert rec.Code, rec.Body.String(), rec.Header()
```

---

## Key terms

- **Handler** — any type implementing `ServeHTTP(http.ResponseWriter,
  *http.Request)`. The fundamental unit of HTTP request handling in Go.
- **HandlerFunc** — a function type with a `ServeHTTP` method that calls itself;
  the adapter that lets a plain function satisfy the `Handler` interface.
- **ServeMux** — Go's built-in request multiplexer (router): maps registered
  method+path patterns to handlers and dispatches the matching one.
- **ResponseWriter** — the interface a handler writes to: `Header()` (set
  headers), `WriteHeader(status)` (send status), `Write([]byte)` (send body),
  in that order.
- **PathValue** — `r.PathValue("name")` reads a `{name}` wildcard segment
  captured by a Go 1.22 ServeMux pattern.
- **DefaultServeMux** — the package-global `ServeMux` used by `http.HandleFunc`
  and a `nil` server handler. Shared mutable state — avoid in production; build
  an explicit mux instead.
- **httptest** — `net/http/httptest`: `NewRecorder` (in-memory ResponseWriter
  for unit-testing handlers) and `NewServer` (a real server on a random local
  port for end-to-end tests).
- **multiplexer** — a router that demultiplexes one stream of incoming requests
  to many handlers based on a pattern; "mux" is short for multiplexer.

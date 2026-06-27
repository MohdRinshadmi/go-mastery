# Day 16 — net/http Deep Dive + Gin Framework

> Mentor note: Today is where Go starts to feel like a real backend language. The standard library's `net/http` package is not a toy — companies like Cloudflare, Tailscale, and countless others run it in production at scale. We'll understand it fully before reaching for Gin. A developer who doesn't understand the stdlib can't debug a framework.

---

## 0. The Big Picture

Before writing a single line, understand what an HTTP server actually is: a program that listens on a TCP port, reads bytes that conform to the HTTP protocol, and writes bytes back. Everything else — routing, middleware, JSON — is built on top of that.

Go's `net/http` package ships with a complete, production-grade HTTP/1.1 and HTTP/2 server. You don't need Nginx as a reverse proxy for many use cases; Go's server IS the server.

```
Browser / curl
      │
      ▼ TCP connection
┌─────────────────────────────┐
│  net/http Server            │
│  ┌─────────────────────┐    │
│  │   ServeMux (router) │    │
│  └────────┬────────────┘    │
│           │ matches pattern  │
│  ┌────────▼────────────┐    │
│  │   http.Handler      │    │
│  │   (your code)       │    │
│  └─────────────────────┘    │
└─────────────────────────────┘
```

---

## 1. The `http.Handler` Interface — The Most Important Interface in Go HTTP

### Theory

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

That's it. Two methods — no, ONE method. Any type that implements `ServeHTTP(ResponseWriter, *Request)` is an HTTP handler. This is Go's philosophy: small interfaces compose into powerful systems.

### Why it exists

Java has `HttpServlet` — a big abstract class you extend. Node has callback functions. Go chose an interface. This means:
- A function can be a handler (via `http.HandlerFunc`)
- A struct can be a handler (stateful handlers)
- Middleware is just a function that takes a Handler and returns a Handler
- You can swap routers without changing your handler code

### The `http.HandlerFunc` adapter

```go
// HandlerFunc is a function type that implements Handler.
// This is the adapter pattern.
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

This lets you write a plain function and convert it:
```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
}

// Register it — http.HandleFunc does the HandlerFunc conversion for you:
http.HandleFunc("/", myHandler)
```

### `ResponseWriter` — what you write to

```go
type ResponseWriter interface {
    Header() http.Header        // set response headers BEFORE writing body
    Write([]byte) (int, error)  // write body (implicitly sends 200 if no WriteHeader)
    WriteHeader(statusCode int) // send status code; must be called before Write
}
```

**Critical ordering rule:** `WriteHeader` → `Header().Set(...)` → `Write()`. Headers must be set before the body. If you call `Write` first, Go sends 200 and locks the headers.

**Senior take:** The most common beginner bug is calling `w.Write(...)` then trying to set a header. By that point the response has started streaming and the header is gone. Always set headers first.

---

## 2. ServeMux — The Standard Library Router

### Theory

`http.ServeMux` is Go's built-in HTTP request multiplexer (router). It matches incoming request paths against registered patterns and dispatches to the right handler.

### Go 1.22 Enhanced Routing (the big upgrade)

Before Go 1.22, ServeMux only matched on paths. You had to parse method and URL params yourself. Go 1.22 added:

```go
mux := http.NewServeMux()

// Method + path pattern:
mux.HandleFunc("GET /products", listProducts)
mux.HandleFunc("POST /products", createProduct)
mux.HandleFunc("GET /products/{id}", getProduct)    // {id} is a named wildcard
mux.HandleFunc("DELETE /products/{id}", deleteProduct)

// In your handler, extract the param:
func getProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")   // new in Go 1.22
    fmt.Fprintf(w, "product %s", id)
}
```

### Pattern matching rules

| Pattern | Matches |
|---------|---------|
| `/` | All paths (catch-all) |
| `/api/` | `/api/` and everything under it (trailing slash = subtree) |
| `/api/users` | Exact match only |
| `GET /api/users` | GET requests to exact path |
| `GET /api/users/{id}` | GET with wildcard segment |
| `GET /api/users/{id...}` | GET with wildcard that matches rest of path |

**When NOT to use stdlib routing:**
- Complex route groups with shared middleware
- OpenAPI spec generation
- Path variables in nested resources: `/orders/{orderID}/items/{itemID}` — works in 1.22 but gets tedious
- When you need automatic OPTIONS, HEAD handling
- Any team with more than ~20 endpoints benefits from a framework

---

## 3. Starting a Server

### Theory

```go
// Method 1: default mux (global state — avoid in production)
http.HandleFunc("/", handler)
http.ListenAndServe(":8080", nil) // nil = use DefaultServeMux

// Method 2: explicit mux (preferred — testable, no global state)
mux := http.NewServeMux()
mux.HandleFunc("GET /", handler)

srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
srv.ListenAndServe()
```

### Why explicit server with timeouts?

The default `http.ListenAndServe` has NO timeouts. A slow or malicious client can hold connections open forever. In production:
- `ReadTimeout`: time to read the entire request including body
- `WriteTimeout`: time to send the response
- `IdleTimeout`: keep-alive connection idle time

**Senior take:** Never ship `http.ListenAndServe(addr, nil)` in production code. It's fine for tutorials. Reviewers will flag it immediately.

---

## 4. Reading Requests

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Method
    method := r.Method  // "GET", "POST", etc.

    // URL parts
    path := r.URL.Path           // "/products/42"
    query := r.URL.Query()       // map[string][]string
    page := query.Get("page")    // "2" or "" if missing

    // Path variables (Go 1.22 ServeMux)
    id := r.PathValue("id")

    // Headers
    contentType := r.Header.Get("Content-Type")
    auth := r.Header.Get("Authorization")

    // Body (JSON decode)
    var payload struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close() // good habit, though ServeHTTP may close it
}
```

---

## 5. Writing Responses

```go
func jsonResponse(w http.ResponseWriter, status int, data any) {
    // Always set Content-Type BEFORE WriteHeader
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Helper for errors:
func jsonError(w http.ResponseWriter, status int, msg string) {
    jsonResponse(w, status, map[string]string{"error": msg})
}
```

### Common mistakes
1. Writing body before headers — headers silently lost
2. Calling `WriteHeader` multiple times — "superfluous response.WriteHeader call" warning
3. Not setting Content-Type — clients guess the type, often wrong
4. Forgetting `return` after error response — handler continues running!

```go
// BUG: handler continues after error
if err != nil {
    http.Error(w, "fail", 500)
    // missing return — code below runs on error!
}
doMoreStuff() // runs even on error path
```

---

## 6. Introducing Gin — When the Framework Earns Its Keep

### Theory

Gin is the most widely used Go HTTP framework. It wraps `net/http` and adds:
- Fast radix-tree router (faster than stdlib for many routes)
- Path param extraction without boilerplate
- Built-in JSON binding and validation
- Middleware chain with `c.Next()` / `c.Abort()`
- Route groups
- Automatic panic recovery

### When to use Gin vs stdlib

| Scenario | Recommendation |
|----------|---------------|
| Simple internal service, < 10 endpoints | stdlib |
| REST API with 20+ endpoints, teams | Gin (or Chi) |
| Need OpenAPI generation | Gin + swaggo |
| Microservice called by known clients | stdlib |
| Public API with input validation needs | Gin |
| Learning / understanding HTTP | stdlib first |

### The `gin.Context` — the heart of Gin

```go
func handler(c *gin.Context) {
    // Path params
    id := c.Param("id")             // /products/:id

    // Query params
    page := c.Query("page")         // ?page=2
    limit := c.DefaultQuery("limit", "10")

    // Bind JSON body
    var req CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Response
    c.JSON(http.StatusOK, gin.H{"id": id, "status": "ok"})
}
```

### Gin vs stdlib comparison

```go
// --- stdlib Go 1.22 ---
mux.HandleFunc("GET /products/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"id": id})
})

// --- Gin equivalent ---
r.GET("/products/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{"id": id})
})
```

Less boilerplate, but notice what Gin hides: you don't control the ResponseWriter directly. That's the tradeoff.

### Route Groups in Gin

```go
api := r.Group("/api/v1")
{
    api.GET("/products", listProducts)
    api.POST("/products", createProduct)

    products := api.Group("/products")
    {
        products.GET("/:id", getProduct)
        products.PUT("/:id", updateProduct)
        products.DELETE("/:id", deleteProduct)
    }
}
```

---

## 7. Performance Implications

- Go's `net/http` server is already concurrent: each request runs in its own goroutine. You get concurrency for free.
- Gin's router uses a radix tree vs ServeMux's linear scan for pre-1.22 patterns. For 1.22+ patterns, stdlib is also efficient.
- Benchmarks matter less than: connection pool config, database latency, JSON encoding. Don't choose a framework based on router benchmarks for typical API work.
- `http.Server` memory: each goroutine uses ~8KB of stack initially. At 100K concurrent connections, that's ~800MB just for stacks. For extreme concurrency, look at `fasthttp` — but that's a different path.

---

## 8. Expert Thinking Mode

- **Beginner:** "I just want to handle a GET request."
- **Intermediate:** "I use Gin because it's easier and has middleware."
- **Senior:** "I understand what Gin is wrapping. I can write my own middleware, I know when `gin.Context` vs `http.ResponseWriter` matters, and I know when stdlib is sufficient."
- **Staff:** "HTTP handler design is the entry point to the system. I care about: timeout configuration, graceful shutdown, the header/body ordering contract, and that my handlers are testable (accept `*http.Request`, return values or write to `ResponseWriter` — and I can test them with `httptest.NewRecorder()`)."

---

## 9. Real-World Use

- **Stripe:** Uses Go's stdlib HTTP server for internal services. Their SDKs expose a `http.Handler` interface so you can plug them into any router.
- **Kubernetes API server:** Heavily customized `net/http` with custom routing. Understanding stdlib is necessary to understand any major Go server.
- **Cloudflare:** Ships Go HTTP servers at the edge. They care deeply about `ReadTimeout`/`WriteTimeout` — a missing timeout is a DoS vector.
- **Our E-Commerce backend:** We'll use Gin for the API layer (Days 17-20) because route groups + middleware chains map cleanly to REST + JWT auth.

---

## 10. Interview Questions

1. What does the `http.Handler` interface look like? Why is it so small?
2. What is `http.HandlerFunc` and what pattern does it demonstrate?
3. What happens if you call `w.Write()` before `w.Header().Set()`?
4. What are the new routing features in Go 1.22's ServeMux?
5. Why should you always set `ReadTimeout` and `WriteTimeout` on `http.Server`?
6. What does `c.Abort()` do in Gin middleware, and how does it differ from just returning?
7. When would you choose stdlib `net/http` over Gin?

---

## Your Tasks for Today

Go to `../exercises/`. You'll build a small product catalog API — first in pure stdlib, then in Gin. Fill in the TODOs and run both servers. I'll review like a real PR.

Don't peek at `../solutions/` until you've tried.

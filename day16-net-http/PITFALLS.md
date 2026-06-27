# Day 16 — net/http Pitfalls (Trap → Why → Fix)

Senior take: every one of these has shipped to production in a real Go service.
They compile. They mostly work. They bite you at the worst time.

---

## 1. Writing the body before the status

**Trap.**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "not found")        // body first
    w.WriteHeader(http.StatusNotFound)  // ignored
}
```

**Why.** The first `Write` implicitly calls `WriteHeader(200)` and freezes the
header. The later `WriteHeader(404)` has nothing to commit — the status line is
already on the wire — so it's dropped and you get a `superfluous
response.WriteHeader call` log. The client checks status `200` and thinks it
succeeded.

**Fix.**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusNotFound) // status first
    fmt.Fprintln(w, "not found")       // then body
}
```

---

## 2. Missing `return` after `http.Error`

**Trap.**

```go
if err := dec.Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    // no return!
}
save(req) // runs even on the error path — and writes a second response
```

**Why.** `http.Error` does not stop your handler. It writes the error response,
but control flow continues. The code below runs with bad data, and any further
`w.Write`/`WriteHeader` triggers `superfluous response.WriteHeader call` and a
garbled, double-bodied response.

**Fix.**

```go
if err := dec.Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return // stop here
}
save(req)
```

Treat "every `http.Error` is immediately followed by `return`" as a hard rule.

---

## 3. `ListenAndServe` with no timeouts (a DoS vector)

**Trap.**

```go
http.ListenAndServe(":8080", mux) // zero timeouts — open forever
```

**Why.** The default server has `ReadTimeout`, `WriteTimeout`, and
`IdleTimeout` all set to zero, meaning *no limit*. A slow client (Slowloris)
can open connections and dribble one byte at a time, holding goroutines and
file descriptors open indefinitely until you run out. This is a real
denial-of-service, not a theoretical one — Cloudflare flags missing timeouts.

**Fix.**

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

---

## 4. Using `DefaultServeMux` / `http.HandleFunc` global state

**Trap.**

```go
func init() { http.HandleFunc("/", index) } // registers on a global
// ...somewhere else, anyone can also do http.Handle("/admin", ...)
http.ListenAndServe(":8080", nil) // nil = DefaultServeMux
```

**Why.** `http.HandleFunc` and `nil` handler both use the package-global
`DefaultServeMux`. It's shared mutable state: any imported package (including a
malicious or careless dependency — see the old `expvar`/`pprof` auto-registration)
can register routes on it. Tests interfere with each other, and you can't run
two independent muxes.

**Fix.**

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /", index)
srv := &http.Server{Addr: ":8080", Handler: mux}
```

Build an explicit `*http.ServeMux` and pass it in. Testable, isolated, no
global state.

---

## 5. Not setting `Content-Type`

**Trap.**

```go
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(data) // no Content-Type set
```

**Why.** If you don't set `Content-Type`, Go sniffs the first 512 bytes and
guesses (`http.DetectContentType`). JSON often gets sniffed as
`text/plain; charset=utf-8`, so browsers and strict clients won't parse it as
JSON. Worse, untyped user-controlled content can be sniffed as `text/html` and
open an XSS hole.

**Fix.**

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(data)
```

Set `Content-Type` explicitly, before `WriteHeader`.

---

## 6. Decoding the request body with no size limit

**Trap.**

```go
var req Payload
json.NewDecoder(r.Body).Decode(&req) // reads an unbounded body
```

**Why.** A client can POST a multi-gigabyte body and your handler will happily
buffer/parse it, exhausting memory. There's also no error checked here, so junk
silently becomes a zero-value `req`. Unbounded reads are a memory-DoS.

**Fix.**

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // cap at 1 MiB
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}
```

`http.MaxBytesReader` caps the read and, combined with checking the decode
error, makes the handler safe. (The server closes `r.Body` for you, but capping
the size is on you.)

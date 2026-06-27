# Day 16 — net/http Resources

Curated, specific to today's material. Start with the routing blog and the
package docs; the Eli Bendersky piece is the best "what actually happens"
read.

- [`net/http` package docs](https://pkg.go.dev/net/http) — the authoritative
  reference for `Handler`, `HandlerFunc`, `ServeMux`, `Server`, and
  `ResponseWriter`. Read the `ResponseWriter` and `ServeMux` doc comments in
  full.
- [Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements) —
  the official blog post introducing method matching, `{wildcard}` segments,
  `PathValue`, and pattern precedence in the stdlib mux.
- [`net/http/httptest` package docs](https://pkg.go.dev/net/http/httptest) —
  `NewRecorder` and `NewServer`; how to test handlers without a real port.
- [Eli Bendersky — "Life of an HTTP request in a Go server"](https://eli.thegreenplace.net/2021/life-of-an-http-request-in-a-go-server/)
  — walks a request from the listener through the per-connection goroutine to
  your handler. The mental model behind "concurrency for free."
- [Alex Edwards — "Let's Go"](https://lets-go.alexedwards.net/) — the canonical
  book for building real web applications with stdlib `net/http` (handlers,
  routing, middleware, sessions). Highly recommended after this day.
- [Gin documentation](https://gin-gonic.com/docs/) — `gin.Context`, route
  groups, `c.JSON`, `ShouldBindJSON`, and the `c.Next()`/`c.Abort()` middleware
  chain we introduced today.

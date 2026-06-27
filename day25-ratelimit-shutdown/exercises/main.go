// Day 25 — YOUR exercises. Run: go run .
package main

import (
	"net/http"
)

// =====================================================================
// TASK 1 — per-IP rate limiter middleware
// Use golang.org/x/time/rate. Keep a map[ip]*rate.Limiter (guard with a
// mutex). On no token: 429 + Retry-After header.
// (go get golang.org/x/time/rate && go mod tidy)
// =====================================================================

// TODO: type ipLimiter struct {...}; newIPLimiter; get(ip); middleware(next)

// =====================================================================
// TASK 2 — http.Server with sane timeouts (ReadTimeout, WriteTimeout,
//          IdleTimeout, ReadHeaderTimeout). NEVER use http.Server{} bare.
// TASK 3 — graceful shutdown:
//   signal.NotifyContext(SIGINT, SIGTERM) -> run server in goroutine ->
//   <-ctx.Done() -> srv.Shutdown(ctxWithTimeout) to drain in-flight.
// =====================================================================

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	// TODO: wrap mux with your rate-limit middleware, build a configured
	// http.Server, run it, and shut it down gracefully on signal.
	http.ListenAndServe(":8080", mux) // replace this with the graceful version
}

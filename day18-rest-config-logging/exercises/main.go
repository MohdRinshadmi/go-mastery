// Day 18 — YOUR exercises. Build a small REST service.
// Run: PORT=8080 go run main.go  then curl localhost:8080/products
package main

import (
	"errors"
	"net/http"
)

// =====================================================================
// EXERCISE 1 — Config with validation + fail-fast
// Finish LoadConfig: read PORT (default 8080) and JWT_SECRET (REQUIRED).
// Return an error if JWT_SECRET is empty so main can exit(1) at startup.
// =====================================================================

type Config struct {
	Port      string
	JWTSecret string
}

func LoadConfig() (Config, error) {
	// TODO: import "os"; read env; require JWT_SECRET
	return Config{}, errors.New("TODO")
}

// =====================================================================
// EXERCISE 2 — consistent JSON helpers
// Implement writeJSON(w, status, v) and writeError(w, status, code, msg)
// that set Content-Type, write the status, and encode JSON. Error shape:
//   {"error": {"code": "...", "message": "..."}}
// =====================================================================

func writeJSON(w http.ResponseWriter, status int, v any) {
	// TODO
}
func writeError(w http.ResponseWriter, status int, code, msg string) {
	// TODO
}

// =====================================================================
// EXERCISE 3 / CHALLENGE — /products endpoint with slog
// - In-memory store with List() and Create(name, price).
// - GET /products -> 200 + JSON array.
// - POST /products -> decode, validate (name non-empty, price>0 -> 422 on
//   fail, 400 on bad JSON), create, log with slog (JSON handler) including
//   a request_id, return 201 + the product.
// Wire slog.SetDefault with a JSON handler in main.
// =====================================================================

func main() {
	// TODO: LoadConfig (exit 1 on error), set up slog JSON logger,
	// build a ServeMux with the two routes, ListenAndServe on cfg.Port.
}

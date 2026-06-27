// Day 22 examples — a Go service with proper CI/CD hooks.
//
// This file demonstrates code that is CI-friendly:
//   - go vet clean
//   - golangci-lint clean
//   - correct error handling (no ignored errors)
//   - no data races
//   - testable structure
//
// Run: go run ./cmd/server
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Version metadata injected by -ldflags in the Makefile / CI pipeline.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// OrderStatus demonstrates a typed constant — linters check exhaustive switches.
type OrderStatus int

const (
	OrderPending   OrderStatus = iota
	OrderPaid
	OrderShipped
	OrderDelivered
)

func (s OrderStatus) String() string {
	switch s {
	case OrderPending:
		return "pending"
	case OrderPaid:
		return "paid"
	case OrderShipped:
		return "shipped"
	case OrderDelivered:
		return "delivered"
	default:
		return "unknown"
	}
}

// parsePort converts an env var to a validated port number.
// This is the kind of function golangci-lint errcheck will verify
// — we must not silently ignore the parse error.
func parsePort(s string, defaultPort int) int {
	if s == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		log.Printf("invalid PORT %q, using default %d", s, defaultPort)
		return defaultPort
	}
	return p
}

// healthHandler is extracted as a named function so it is unit-testable.
// CI runs `go test -race ./...` which will test this.
func healthHandler(version, gitCommit string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":     "ok",
			"version":    version,
			"git_commit": gitCommit,
		}); err != nil {
			// errcheck linter requires we handle this.
			// In a real service, log.Printf("health encode: %v", err) and return.
			log.Printf("health encode error: %v", err)
		}
	}
}

func main() {
	port := parsePort(os.Getenv("PORT"), 8080)

	mux := http.NewServeMux()
	mux.Handle("GET /health", healthHandler(Version, GitCommit))
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"version":    Version,
			"git_commit": GitCommit,
			"build_time": BuildTime,
		}); err != nil {
			log.Printf("version encode error: %v", err)
		}
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("server starting on :%d (version=%s commit=%s)", port, Version, GitCommit)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

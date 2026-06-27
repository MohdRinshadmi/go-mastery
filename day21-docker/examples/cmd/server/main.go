// Day 21 examples — production-ready HTTP server skeleton to containerize.
//
// Build:   go build -o bin/server ./cmd/server
// Run:     ./bin/server
// Docker:  docker build -t day21-shop-api . && docker run -p 8080:8080 day21-shop-api
//
// This is the service we will containerize. It models a minimal e-commerce API
// that would exist after Phase 4. The focus today is the Dockerfile / Compose,
// not the service logic.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Version metadata — injected at build time via -ldflags.
// When you build with:
//
//	go build -ldflags="-X main.Version=v1.2.3 -X main.GitCommit=abc1234" ...
//
// these variables are overwritten in the final binary. The defaults ("dev",
// "unknown") are only used during local `go run`.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// config holds everything the service needs from the environment.
// Twelve-factor apps read configuration from the environment, not from files.
type config struct {
	port        string
	databaseURL string
	redisURL    string
}

func loadConfig() config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return config{
		port:        port,
		databaseURL: os.Getenv("DATABASE_URL"),
		redisURL:    os.Getenv("REDIS_URL"),
	}
}

func main() {
	cfg := loadConfig()

	mux := http.NewServeMux()

	// GET /health — liveness probe (Day 24 will expand this significantly)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// GET /version — tells you exactly what's deployed. Invaluable in prod.
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"version":    Version,
			"git_commit": GitCommit,
			"build_time": BuildTime,
		})
	})

	// GET /config — shows resolved config (never expose secrets in real apps!)
	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// In real services, NEVER expose raw config over HTTP.
		// This is only for teaching: showing that env vars are injected.
		dbURL := cfg.databaseURL
		if dbURL == "" {
			dbURL = "(not set)"
		}
		redisURL := cfg.redisURL
		if redisURL == "" {
			redisURL = "(not set)"
		}
		json.NewEncoder(w).Encode(map[string]string{
			"database_url": dbURL,
			"redis_url":    redisURL,
		})
	})

	// GET /products — simulated product listing
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		type Product struct {
			ID    int     `json:"id"`
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}
		products := []Product{
			{1, "Laptop", 999.99},
			{2, "Mouse", 29.99},
			{3, "Keyboard", 79.99},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("server starting on :%s (version=%s commit=%s)", cfg.port, Version, GitCommit)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

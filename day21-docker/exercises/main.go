// Day 21 — YOUR exercises. Fill in the TODOs.
//
// This file is the Go service you'll containerize. It already runs — your
// job is to write the Dockerfile and Compose files.
//
// Run locally:    go run main.go
// Then: write Dockerfile, .dockerignore, and docker-compose.yml
// See the TODO comments below for what to build.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// =====================================================================
// EXERCISE 1 — Write the Dockerfile
//
// In this directory, create a file called Dockerfile that:
//   1. Uses golang:1.22-alpine as the builder stage
//   2. Copies go.mod (and go.sum if it exists) BEFORE copying source
//      (why? layer caching — see lesson section 1)
//   3. Builds with CGO_ENABLED=0 and -ldflags="-s -w"
//   4. Uses gcr.io/distroless/static:nonroot as the runtime stage
//   5. Runs as the nonroot user
//   6. Exposes port 8080
//
// Verify with: docker build -t ex21 . && docker run -p 8080:8080 ex21
// Then curl http://localhost:8080/health and check the response.
// =====================================================================

// =====================================================================
// EXERCISE 2 — Debug the broken Dockerfile
//
// Open broken_dockerfile.txt (in this directory). It has 4 mistakes.
// Find them all and write the fixed Dockerfile as fixed_dockerfile.txt.
//
// Hint: think about CGO, user, CA certs, and layer ordering.
// =====================================================================

// =====================================================================
// EXERCISE 3 — Write the docker-compose.yml
//
// Create docker-compose.yml that runs this service + Postgres + Redis.
// Requirements:
//   - api service builds from this Dockerfile
//   - postgres service uses postgres:16-alpine with a healthcheck
//   - api depends_on postgres with condition: service_healthy
//   - redis service uses redis:7-alpine
//   - named volumes for pgdata and redisdata
//   - environment variables: DATABASE_URL and REDIS_URL pointing at the services
//
// Verify: docker compose up -d, then:
//   curl http://localhost:8080/health   -> {"status":"ok"}
//   curl http://localhost:8080/config   -> shows DATABASE_URL and REDIS_URL
// =====================================================================

// =====================================================================
// CHALLENGE — Add version injection
//
// Modify your Dockerfile to accept VERSION, GIT_COMMIT, BUILD_TIME as
// build arguments and inject them via -ldflags -X. Then:
//   docker build --build-arg VERSION=v2.0.0 --build-arg GIT_COMMIT=abc123 -t ex21-v2 .
//   docker run -p 8080:8080 ex21-v2
//   curl http://localhost:8080/version   -> {"version":"v2.0.0","git_commit":"abc123",...}
// =====================================================================

// Version metadata (injected at build time — see CHALLENGE above).
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"version":    Version,
			"git_commit": GitCommit,
			"build_time": BuildTime,
		})
	})

	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "(not set)"
		}
		redisURL := os.Getenv("REDIS_URL")
		if redisURL == "" {
			redisURL = "(not set)"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"database_url": dbURL,
			"redis_url":    redisURL,
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

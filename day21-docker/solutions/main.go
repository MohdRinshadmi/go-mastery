// Day 21 — solutions reference. Try the exercises yourself FIRST.
// This is the same service as exercises/main.go — the solutions are
// in the Docker artifacts below, not in Go code.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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

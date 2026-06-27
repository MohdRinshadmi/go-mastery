// Day 18 — reference solution. Run: JWT_SECRET=dev PORT=8080 go run main.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

type Config struct {
	Port      string
	JWTSecret string
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:      getenv("PORT", "8080"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
type CreateReq struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (r CreateReq) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Price <= 0 {
		return errors.New("price must be > 0")
	}
	return nil
}

type Store struct {
	mu    sync.Mutex
	items []Product
	next  int
}

func (s *Store) List() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Product(nil), s.items...)
}
func (s *Store) Create(name string, price float64) Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	p := Product{ID: s.next, Name: name, Price: price}
	s.items = append(s.items, p)
	return p
}

var reqCounter atomic.Int64

// middleware: attach a request_id-scoped logger to each request via context.
func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("req-%d", reqCounter.Add(1))
		log := slog.With("request_id", id, "method", r.Method, "path", r.URL.Path)
		log.Info("request received")
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	store := &Store{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.List())
	})
	mux.HandleFunc("POST /products", func(w http.ResponseWriter, r *http.Request) {
		var req CreateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		if err := req.Validate(); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
			return
		}
		p := store.Create(req.Name, req.Price)
		slog.Info("product created", "product_id", p.ID)
		writeJSON(w, http.StatusCreated, p)
	})

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, withRequestLog(mux)); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

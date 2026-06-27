// Day 18 walkthrough — REST + config + slog. Run: PORT=8080 go run main.go
// Then: curl localhost:8080/products ; curl -XPOST localhost:8080/products -d '{"name":"Pen","price":2.5}'
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

// ---- Config (12-factor, validated at startup) ---------------------------
type Config struct {
	Port     string
	LogLevel string
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:     getenv("PORT", "8080"),
		LogLevel: getenv("LOG_LEVEL", "info"),
	}
	if cfg.Port == "" {
		return cfg, errors.New("PORT required")
	}
	return cfg, nil
}

// ---- Domain + in-memory store -------------------------------------------
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (r CreateProductRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Price <= 0 {
		return errors.New("price must be > 0")
	}
	return nil
}

type Store struct {
	mu     sync.Mutex
	items  []Product
	nextID int
}

func (s *Store) List() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Product(nil), s.items...)
}
func (s *Store) Create(name string, price float64) Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	p := Product{ID: s.nextID, Name: name, Price: price}
	s.items = append(s.items, p)
	return p
}

// ---- HTTP helpers: consistent JSON + error shape ------------------------
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

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		// fail fast — don't start in a broken state
		slog.Error("config error", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	store := &Store{}

	mux := http.NewServeMux()
	// Go 1.22 method+pattern routing
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.List())
	})
	mux.HandleFunc("POST /products", func(w http.ResponseWriter, r *http.Request) {
		var req CreateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		if err := req.Validate(); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
			return
		}
		p := store.Create(req.Name, req.Price)
		slog.Info("product created", "product_id", p.ID, "name", p.Name)
		writeJSON(w, http.StatusCreated, p)
	})

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

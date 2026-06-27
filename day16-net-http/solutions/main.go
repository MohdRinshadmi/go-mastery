// Day 16 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
)

// ---- Exercise 1: HealthHandler ---------------------------------------------

type HealthHandler struct{}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "catalog",
	})
}

// ---- Exercise 2: Category routes -------------------------------------------

func runCategoryServer() http.Handler {
	mux := http.NewServeMux()

	categories := []string{"Electronics", "Books", "Clothing"}

	mux.HandleFunc("GET /api/categories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(categories)
	})

	mux.HandleFunc("GET /api/categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":   id,
			"name": "Electronics",
		})
	})

	return mux
}

// ---- Exercise 3: Order router ----------------------------------------------

type CreateOrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func buildOrderRouter() *gin.Engine {
	// Suppress Gin's startup banner and debug logs for clean output.
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	{
		api.POST("/orders", func(c *gin.Context) {
			var req CreateOrderRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
				return
			}
			// Manual validation — explicit, readable, no magic tags yet.
			if req.ProductID == "" {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "product_id is required"})
				return
			}
			if req.Quantity <= 0 {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "quantity must be > 0"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"order_id": "ord-001",
				"status":   "pending",
			})
		})

		api.GET("/orders/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"order_id": id,
				"status":   "pending",
			})
		})
	}

	return r
}

// ---- Challenge: Middleware chain --------------------------------------------

// onlyJSON enforces JSON Content-Type on mutating methods.
// The key insight: this IS an http.Handler (via HandlerFunc), and it takes
// an http.Handler — so it composes cleanly.
func onlyJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutating := r.Method == http.MethodPost ||
			r.Method == http.MethodPut ||
			r.Method == http.MethodPatch

		if mutating && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Content-Type must be application/json",
			})
			return // CRITICAL: stop the chain here
		}
		next.ServeHTTP(w, r)
	})
}

// requestLogger prints method + path before passing to next.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("  [%s] %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// --- Exercise 1 ---
	fmt.Println("== Exercise 1: HealthHandler ==")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	HealthHandler{}.ServeHTTP(rec, req)
	fmt.Printf("  status: %d\n  body:   %s", rec.Code, rec.Body.String())

	// --- Exercise 2 ---
	fmt.Println("== Exercise 2: Category routes ==")
	handler := runCategoryServer()

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	handler.ServeHTTP(rec, req)
	fmt.Printf("  GET /api/categories → %d: %s", rec.Code, rec.Body.String())

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/categories/7", nil)
	handler.ServeHTTP(rec, req)
	fmt.Printf("  GET /api/categories/7 → %d: %s", rec.Code, rec.Body.String())

	// --- Exercise 3 ---
	fmt.Println("== Exercise 3: Order router ==")
	router := buildOrderRouter()

	// Valid order
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/orders",
		strings.NewReader(`{"product_id":"p-1","quantity":3}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /api/v1/orders (valid) → %d: %s", rec.Code, rec.Body.String())

	// Missing product_id
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/orders",
		strings.NewReader(`{"quantity":3}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /api/v1/orders (no product_id) → %d: %s", rec.Code, rec.Body.String())

	// GET order
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/orders/ord-99", nil)
	router.ServeHTTP(rec, req)
	fmt.Printf("  GET /api/v1/orders/ord-99 → %d: %s", rec.Code, rec.Body.String())

	// --- Challenge ---
	fmt.Println("== Challenge: Middleware chain ==")

	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	chain := requestLogger(onlyJSON(echoHandler))

	// JSON body → should pass
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/echo", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	chain.ServeHTTP(rec, req)
	fmt.Printf("  POST with JSON → %d: %s", rec.Code, rec.Body.String())

	// Wrong Content-Type → should be 415
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/echo", strings.NewReader("hello"))
	req.Header.Set("Content-Type", "text/plain")
	chain.ServeHTTP(rec, req)
	fmt.Printf("  POST with text/plain → %d: %s", rec.Code, rec.Body.String())

	// GET → should pass (not a mutating method)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	chain.ServeHTTP(rec, req)
	fmt.Printf("  GET (no body) → %d: %s", rec.Code, rec.Body.String())
}

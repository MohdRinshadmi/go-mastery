// Day 16 examples — net/http deep dive + Gin
//
// This file demonstrates both stdlib and Gin patterns side by side.
// It starts an HTTP server — run with: go run main.go
// Then in another terminal:
//   curl http://localhost:8080/api/stdlib/products
//   curl http://localhost:8080/api/gin/products
//   curl -X POST http://localhost:8080/api/gin/products \
//        -H "Content-Type: application/json" \
//        -d '{"name":"Widget","price":9.99}'
//   curl http://localhost:8080/api/gin/products/42
//
// Stop the server with Ctrl+C.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// ---- Domain types ----------------------------------------------------------

// Product is our tiny domain model. JSON tags control serialization names.
type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CreateProductRequest is the shape of the POST body.
// Notice it is separate from Product — the client doesn't set the ID.
type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ---- In-memory "database" --------------------------------------------------

// products is a package-level slice. In production this would be a repository
// interface backed by Postgres (see Day 19). For today, it keeps examples simple.
var products = []Product{
	{ID: "1", Name: "Wireless Mouse", Price: 29.99},
	{ID: "2", Name: "Mechanical Keyboard", Price: 89.99},
	{ID: "3", Name: "USB-C Hub", Price: 49.99},
}

// ---- stdlib helpers --------------------------------------------------------

// writeJSON is a reusable helper — every production stdlib server has one.
// Rule: set headers BEFORE calling WriteHeader, WriteHeader BEFORE Write.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// At this point we've already sent headers — we can't send a new status.
		// Log the error; the partial response is already gone.
		log.Printf("writeJSON encode error: %v", err)
	}
}

// writeJSONError sends a consistent error envelope.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- stdlib handlers -------------------------------------------------------

// stdlibListProducts handles GET /api/stdlib/products
// It demonstrates the raw http.HandlerFunc signature.
func stdlibListProducts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, products)
}

// stdlibGetProduct handles GET /api/stdlib/products/{id}
// Go 1.22 feature: {id} in the pattern + r.PathValue("id")
func stdlibGetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Go 1.22 — zero boilerplate path extraction

	for _, p := range products {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeJSONError(w, http.StatusNotFound, fmt.Sprintf("product %s not found", id))
}

// stdlibCreateProduct handles POST /api/stdlib/products
func stdlibCreateProduct(w http.ResponseWriter, r *http.Request) {
	// Always check Content-Type in production — defend your API.
	if r.Header.Get("Content-Type") != "application/json" {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req CreateProductRequest
	// Limit body size to prevent memory exhaustion attacks.
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) // 1 MB
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	defer r.Body.Close()

	if req.Name == "" {
		writeJSONError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	// In production: call repository.Create(ctx, req), get back ID from DB.
	newID := fmt.Sprintf("%d", len(products)+1)
	p := Product{ID: newID, Name: req.Name, Price: req.Price}
	products = append(products, p)

	writeJSON(w, http.StatusCreated, p)
}

// ---- Building the stdlib mux -----------------------------------------------

// newStdlibMux creates a mux with Go 1.22 method+pattern routing.
// Returns an http.Handler so it can be mounted as a sub-tree.
func newStdlibMux() http.Handler {
	mux := http.NewServeMux()

	// Go 1.22: "METHOD /path" syntax — clean, built-in.
	mux.HandleFunc("GET /products", stdlibListProducts)
	mux.HandleFunc("POST /products", stdlibCreateProduct)
	mux.HandleFunc("GET /products/{id}", stdlibGetProduct)

	// Strip the /api/stdlib prefix before passing to this mux.
	// http.StripPrefix is the stdlib way to mount at a sub-path.
	return http.StripPrefix("/api/stdlib", mux)
}

// ---- Gin handlers ----------------------------------------------------------

// ginListProducts is the Gin equivalent of stdlibListProducts.
// Notice: c.JSON handles Content-Type, WriteHeader, and Encode in one call.
func ginListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, products)
}

// ginGetProduct uses Gin's :id path param syntax.
func ginGetProduct(c *gin.Context) {
	id := c.Param("id") // note: "id" not ":id"

	for _, p := range products {
		if p.ID == id {
			c.JSON(http.StatusOK, p)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("product %s not found", id)})
}

// ginCreateProduct uses ShouldBindJSON for decode + validation in one step.
func ginCreateProduct(c *gin.Context) {
	var req CreateProductRequest

	// ShouldBindJSON decodes the body and returns an error on failure.
	// (Use MustBindJSON if you want automatic 400 on failure — but that
	// calls c.AbortWithStatus, which prevents custom error envelopes.)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "name is required"})
		return
	}

	newID := fmt.Sprintf("%d", len(products)+1)
	p := Product{ID: newID, Name: req.Name, Price: req.Price}
	products = append(products, p)

	c.JSON(http.StatusCreated, p)
}

// ---- Building the Gin engine -----------------------------------------------

// newGinEngine builds a configured Gin engine with our routes.
// Returning *gin.Engine (which implements http.Handler) lets us embed Gin
// inside a stdlib http.Server for full timeout control.
func newGinEngine() *gin.Engine {
	// gin.New() is the production form — no default logger/recovery.
	// gin.Default() adds Logger + Recovery middleware automatically.
	// We use Default here for learning; in Day 17 we'll add custom middleware.
	r := gin.Default()

	// Route group: all product routes share the /api/gin/products prefix.
	api := r.Group("/api/gin")
	{
		products := api.Group("/products")
		{
			products.GET("", ginListProducts)
			products.POST("", ginCreateProduct)
			products.GET("/:id", ginGetProduct)
		}
	}

	return r
}

// ---- Combining both under one server ---------------------------------------

// combinedHandler routes /api/stdlib/* to stdlib and /api/gin/* to Gin.
// This is purely for demo purposes — real apps use one or the other.
func combinedHandler(stdlibH, ginH http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/stdlib/", stdlibH)
	mux.Handle("/api/gin/", ginH)

	// Root info endpoint
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"stdlib_products": "GET /api/stdlib/products",
			"gin_products":    "GET /api/gin/products",
			"gin_product":     "GET /api/gin/products/:id",
		})
	})

	return mux
}

// ---- main: production-grade server with graceful shutdown ------------------

func main() {
	// Build handlers
	stdlibH := newStdlibMux()
	ginH := newGinEngine()
	handler := combinedHandler(stdlibH, ginH)

	// Explicit http.Server with timeouts — never use ListenAndServe bare.
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine so we can listen for OS signals.
	go func() {
		fmt.Println("Server listening on :8080")
		fmt.Println("Try: curl http://localhost:8080/api/gin/products")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown: wait for SIGINT or SIGTERM.
	// In production (Docker/K8s) the orchestrator sends SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown waits for in-flight requests to finish, up to the timeout.
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	fmt.Println("Server stopped.")
}

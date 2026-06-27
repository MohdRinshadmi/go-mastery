// Day 20 — Debugging Challenge (FIXED): the composition root wires every dependency.
//
// Same compact clean-architecture skeleton as ../bugged, with ONE change in
// main(): we construct the in-memory repository and inject it into the service.
// Now the service can reach data, so an existing product returns 200 and a
// missing one is mapped to 404 by the single statusFor() function.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// ----- domain (innermost: imports nothing) -----

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("internal error")
)

type Product struct {
	ID   string
	Name string
}

// ----- repository (interface owned by the service; in-memory implementation) -----

type ProductRepo interface {
	FindByID(ctx context.Context, id string) (Product, error)
}

type inMemoryRepo struct {
	products map[string]Product
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{products: map[string]Product{
		"p1": {ID: "p1", Name: "Fountain Pen"},
	}}
}

func (r *inMemoryRepo) FindByID(_ context.Context, id string) (Product, error) {
	p, ok := r.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return p, nil
}

// ----- service (depends on the repo INTERFACE, never net/http or a DB) -----

type ProductService struct {
	repo ProductRepo
}

func NewService(repo ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Get(ctx context.Context, id string) (Product, error) {
	if s.repo == nil {
		return Product{}, fmt.Errorf("product service has no repository: %w", ErrInternal)
	}
	return s.repo.FindByID(ctx, id)
}

// ----- transport/http (thin: call service, map domain error -> status) -----

// statusFor is the ONE place domain errors become HTTP status codes.
func statusFor(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

type productHandler struct {
	svc *ProductService
}

func (h *productHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	p, err := h.svc.Get(r.Context(), id)
	w.WriteHeader(statusFor(err))
	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
		return
	}
	fmt.Fprintf(w, "product: %s", p.Name)
}

// ----- composition root -----

func main() {
	// FIX: construct the concrete repo and inject it. The composition root is
	// the ONLY place that knows the concrete wiring; every dependency is wired.
	repo := newInMemoryRepo()
	svc := NewService(repo)
	h := &productHandler{svc: svc}

	fmt.Println("=== fixed ===")

	// Existing product -> 200.
	reqOK := httptest.NewRequest(http.MethodGet, "/products?id=p1", nil)
	recOK := httptest.NewRecorder()
	h.get(recOK, reqOK)
	fmt.Printf("GET /products?id=p1      -> status %d (want 200)\n", recOK.Code)
	fmt.Printf("body: %s\n", recOK.Body.String())

	// Missing product -> 404, mapped by statusFor from the domain ErrNotFound.
	reqMiss := httptest.NewRequest(http.MethodGet, "/products?id=nope", nil)
	recMiss := httptest.NewRecorder()
	h.get(recMiss, reqMiss)
	fmt.Printf("GET /products?id=nope    -> status %d (want 404)\n", recMiss.Code)
	fmt.Printf("body: %s\n", recMiss.Body.String())
}

// Day 20 — Debugging Challenge (BUGGED): nil dependency at the composition root.
//
// A compact clean-architecture skeleton in ONE file. We simulate the four
// layers with distinct types:
//
//	domain      -> Product + domain errors (ErrNotFound, ...)
//	repository  -> ProductRepo INTERFACE + inMemoryRepo implementation
//	service     -> ProductService depends on the repo INTERFACE, not the impl
//	transport   -> productHandler turns HTTP into a service call, maps errors
//	              with the ONE statusFor() function
//
// THE BUG lives in the composition root (main): it builds the service with a
// nil repository — NewService(nil) — forgetting to inject the in-memory repo.
// The service guards against a nil repo and returns a domain error, so instead
// of 200 (product exists) we get 500. Deterministic via httptest, no live port.
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
	// Guard: if wiring is wrong the repo is nil. Rather than panic on a nil
	// dereference, surface it as a domain error so the boundary maps it to 500.
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
	// BUG: forgot to construct and inject the repository. The repo is nil, so
	// the service can never reach data and every request fails with 500.
	svc := NewService(nil)
	h := &productHandler{svc: svc}

	// Deterministic demonstration with httptest — no real port.
	req := httptest.NewRequest(http.MethodGet, "/products?id=p1", nil)
	rec := httptest.NewRecorder()
	h.get(rec, req)

	fmt.Println("=== bugged ===")
	fmt.Printf("GET /products?id=p1  -> status %d (want 200)\n", rec.Code)
	fmt.Printf("body: %s\n", rec.Body.String())
}

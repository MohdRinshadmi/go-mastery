package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetProduct_MissingReturns404 locks in the header/body ordering contract:
// because the handler calls WriteHeader(404) before writing the body, the
// recorder reports 404 — not the implicit 200 the bugged version produces.
func TestGetProduct_MissingReturns404(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/products/missing", nil)
	req.SetPathValue("id", "missing")

	getProduct(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

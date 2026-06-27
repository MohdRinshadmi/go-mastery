package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// getProduct looks up a product by its path value and returns 404 when the
// product is missing.
//
// FIX: set headers and the status line BEFORE writing the body. WriteHeader
// commits the chosen status; the subsequent Write only appends body bytes.
// Order is: Header().Set -> WriteHeader(status) -> Write(body).
func getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "missing" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound) // status committed first
		fmt.Fprintf(w, "product %q not found\n", id)
		return
	}
	w.Write([]byte("product found\n"))
}

func main() {
	// Deterministic demonstration: no live port, no flakiness.
	// httptest.NewRecorder captures exactly what the handler wrote.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/products/missing", nil)
	req.SetPathValue("id", "missing")

	getProduct(rec, req)

	fmt.Printf("=== fixed ===\n")
	fmt.Printf("intended status=404, got status=%d\n", rec.Code)
}

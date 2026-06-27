package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// getProduct looks up a product by its path value. When the product is
// missing it intends to return 404 Not Found.
//
// BUG: it writes the response BODY first, then calls w.WriteHeader(404).
// The first Write implicitly commits a 200 OK status and locks the header.
// By the time WriteHeader runs the status line is already on the wire, so
// the 404 is silently dropped (you also get a "superfluous response.WriteHeader
// call" log line from a live server). The client sees 200 with the error body.
func getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "missing" {
		// Body goes out FIRST -> commits 200 and freezes the header.
		fmt.Fprintf(w, "product %q not found\n", id)
		w.WriteHeader(http.StatusNotFound) // too late: already 200
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

	fmt.Printf("=== bugged ===\n")
	fmt.Printf("intended status=404, got status=%d\n", rec.Code)
}

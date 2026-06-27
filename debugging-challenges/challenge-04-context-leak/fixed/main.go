package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func queryDB(ctx context.Context, q string) (string, error) {
	select {
	case <-time.After(2 * time.Second):
		return "rows for: " + q, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// FIX: derive from r.Context() so client disconnects propagate downstream.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := queryDB(ctx, "SELECT * FROM orders")
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		fmt.Println("handler: aborted ->", err)
		return
	}
	fmt.Println("handler: finished work for a client that may be gone")
	fmt.Fprintln(w, result)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	handler(rec, req)
	elapsed := time.Since(start)

	fmt.Printf("handler returned after %v (want ~50ms, client left early)\n", elapsed.Round(time.Millisecond))
}

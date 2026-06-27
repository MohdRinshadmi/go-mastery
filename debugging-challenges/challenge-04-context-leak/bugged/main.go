package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// queryDB simulates a slow downstream call. It respects its context — so the
// ONLY reason it won't cancel here is that the wrong context is plumbed into it.
func queryDB(ctx context.Context, q string) (string, error) {
	select {
	case <-time.After(2 * time.Second): // the "real" work
		return "rows for: " + q, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// BUG: context derived from Background(), not from r.Context().
	// When the client disconnects, r.Context() is cancelled — but this ctx
	// never hears about it, so queryDB grinds through the full 2s.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	// Build a request whose context we cancel after 50ms to simulate the
	// client disconnecting.
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel() // client goes away
	}()

	start := time.Now()
	handler(rec, req)
	elapsed := time.Since(start)

	fmt.Printf("handler returned after %v (want ~50ms, client left early)\n", elapsed.Round(time.Millisecond))
}

// Day 24 — YOUR exercises. Run: go run .
package main

import (
	"context"
	"net/http"
	"sync/atomic"
)

var dependencyReady atomic.Bool

// =====================================================================
// TASK 1 — /healthz liveness: return 200 always (no dependency checks!).
// TASK 2 — /readyz readiness: check checkDependency with a 2s timeout;
//          503 if not ready, 200 if ready.
// =====================================================================

func checkDependency(ctx context.Context) error {
	// TODO: simulate a dependency check that respects ctx (select + time.After)
	return nil
}

// =====================================================================
// CHALLENGE — trace id through context into logs
// Implement withTrace(ctx, id) / traceID(ctx) and a /order handler that
// generates a trace id, stores it in ctx, and logs two "spans" (placeOrder,
// charge) that both include the SAME trace_id (pulled from ctx).
// =====================================================================

func main() {
	dependencyReady.Store(true)
	mux := http.NewServeMux()

	// TODO TASK 1
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
	// TODO TASK 2
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
	// TODO CHALLENGE: /order

	http.ListenAndServe(":8080", mux)
}

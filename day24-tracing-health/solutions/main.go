// Day 24 walkthrough — health checks + a tiny context-propagated tracer.
// Run: go run .   then: curl localhost:8080/healthz ; curl localhost:8080/readyz
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// ---- mini-tracer: shows the IDEA behind OTel (trace id in ctx + spans) ---
type ctxKey string

const traceKey ctxKey = "trace_id"

var spanCounter atomic.Int64

func withTrace(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey, id)
}
func traceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceKey).(string); ok {
		return v
	}
	return "none"
}

// startSpan returns an end() func that logs duration — a toy span.
func startSpan(ctx context.Context, name string) func() {
	start := time.Now()
	n := spanCounter.Add(1)
	return func() {
		slog.Info("span",
			"trace_id", traceID(ctx),
			"span", name,
			"span_id", n,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}
}

func placeOrder(ctx context.Context) {
	defer startSpan(ctx, "placeOrder")()
	charge(ctx)
	time.Sleep(5 * time.Millisecond)
}
func charge(ctx context.Context) {
	defer startSpan(ctx, "charge")() // child span, same trace_id via ctx
	time.Sleep(15 * time.Millisecond)
}

// ---- readiness: simulated dependency check with timeout -----------------
var dependencyReady atomic.Bool

func checkDependency(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Millisecond):
		if !dependencyReady.Load() {
			return fmt.Errorf("dependency not ready")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	dependencyReady.Store(true) // pretend startup finished

	mux := http.NewServeMux()

	// Liveness: dumb — do NOT check dependencies here.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive\n"))
	})

	// Readiness: check the critical dependency with a timeout.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := checkDependency(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready\n"))
	})

	// A traced endpoint: a trace_id flows through ctx into every span + log.
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		ctx := withTrace(r.Context(), fmt.Sprintf("trace-%d", time.Now().UnixNano()%100000))
		placeOrder(ctx)
		w.Write([]byte("ordered (see span logs)\n"))
	})

	slog.Info("listening", "port", "8080")
	http.ListenAndServe(":8080", mux)
}

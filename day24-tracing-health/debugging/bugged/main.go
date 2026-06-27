// Day 24 debugging — liveness that checks the database (restart storm).
//
// The service exposes /healthz (liveness) and /readyz (readiness). The
// author wired BOTH to ping the database. When the DB has a transient blip,
// liveness fails → Kubernetes RESTARTS every pod simultaneously → they all
// reconnect at once (thundering herd) → the DB falls over for good →
// CrashLoopBackOff. A DB hiccup that should have been a brief readiness
// outage becomes a full outage.
//
// STDLIB ONLY, no real DB and no real server: we use httptest and a
// togglable fake dependency, then exit promptly.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// db simulates a dependency that can be up or down.
type db struct{ up bool }

func (d *db) Ping() error {
	if !d.up {
		return fmt.Errorf("db: connection refused")
	}
	return nil
}

func newMux(database *db) *http.ServeMux {
	mux := http.NewServeMux()

	// Liveness — "is the process alive?"
	// BUG: it checks the database. A DB outage now makes liveness fail,
	// which tells the orchestrator to RESTART the pod. That's the restart
	// storm anti-pattern.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Readiness — "can I serve traffic right now?" Checking the DB here is
	// correct.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func probe(mux http.Handler, path string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr.Code
}

func main() {
	database := &db{up: true}
	mux := newMux(database)

	fmt.Println("== DB healthy ==")
	fmt.Printf("  /healthz -> %d   /readyz -> %d\n", probe(mux, "/healthz"), probe(mux, "/readyz"))

	// Simulate a transient DB blip.
	database.up = false
	fmt.Println("== DB has a transient blip ==")
	live := probe(mux, "/healthz")
	ready := probe(mux, "/readyz")
	fmt.Printf("  /healthz -> %d   /readyz -> %d\n", live, ready)

	if live != http.StatusOK {
		fmt.Println("=> BUG: liveness FAILED on a DB blip -> orchestrator restarts every pod (restart storm)")
	}
}

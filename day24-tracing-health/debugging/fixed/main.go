// Day 24 debugging — FIXED.
//
// Split the two probes by responsibility:
//   - Liveness (/healthz) = "is the process alive / not deadlocked?" — a dumb
//     200. It must NOT depend on downstreams, so a DB blip can never trigger
//     a restart.
//   - Readiness (/readyz) = "can I serve right now?" — this is where the DB
//     check belongs. On a blip, the pod is pulled OUT OF ROTATION (no new
//     traffic) but NOT restarted; when the DB recovers, readiness goes green
//     and traffic resumes. No restart storm.
//
// STDLIB ONLY, httptest + togglable fake dependency, exits promptly.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type db struct{ up bool }

func (d *db) Ping() error {
	if !d.up {
		return fmt.Errorf("db: connection refused")
	}
	return nil
}

func newMux(database *db) *http.ServeMux {
	mux := http.NewServeMux()

	// Liveness: dumb 200. If the process can run this handler, it's alive.
	// FIX: no dependency checks here.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Readiness: check critical dependencies. Failing here pulls the pod from
	// rotation without restarting it.
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

	database.up = false
	fmt.Println("== DB has a transient blip ==")
	live := probe(mux, "/healthz")
	ready := probe(mux, "/readyz")
	fmt.Printf("  /healthz -> %d   /readyz -> %d\n", live, ready)

	if live == http.StatusOK && ready == http.StatusServiceUnavailable {
		fmt.Println("=> CORRECT: liveness stays 200 (no restart); readiness 503 (pulled from rotation)")
	}

	database.up = true
	fmt.Println("== DB recovers ==")
	fmt.Printf("  /healthz -> %d   /readyz -> %d  (traffic resumes, no restart needed)\n",
		probe(mux, "/healthz"), probe(mux, "/readyz"))
}

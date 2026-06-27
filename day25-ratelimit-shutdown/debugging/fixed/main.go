// Day 25 debugging — FIXED.
//
// Replace srv.Close() with srv.Shutdown(ctx). Shutdown:
//   1. stops accepting NEW connections,
//   2. WAITS for in-flight requests to finish, up to the ctx deadline,
//   3. returns when drained (or the deadline forces it).
//
// The in-flight /work request now completes with 200 instead of being
// severed — that's the difference between a zero-downtime deploy and a blip
// of 5xx. We also WAIT for Shutdown to return (and give it a deadline) so we
// don't exit mid-drain.
//
// STDLIB ONLY. Real server on a random localhost port; exits promptly.
//
// Run with: go run -race .
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "done")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)

	url := "http://" + ln.Addr().String() + "/work"
	client := &http.Client{Timeout: 2 * time.Second}

	var wg sync.WaitGroup
	var reqErr error
	var status int

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := client.Get(url)
		if err != nil {
			reqErr = err
			return
		}
		defer resp.Body.Close()
		status = resp.StatusCode
	}()

	time.Sleep(50 * time.Millisecond) // let the request get in-flight

	// FIX: graceful shutdown drains in-flight requests within the deadline.
	// The deadline must be shorter than the orchestrator's grace period so we
	// finish before SIGKILL.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("forced shutdown:", err)
	}

	wg.Wait()

	if reqErr != nil {
		fmt.Printf("in-flight request FAILED: %v\n", reqErr)
		return
	}
	fmt.Printf("in-flight request completed with status %d\n", status)
	if status == http.StatusOK {
		fmt.Println("=> CORRECT: srv.Shutdown drained the in-flight request — zero-downtime deploy")
	}
}

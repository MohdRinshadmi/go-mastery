// Day 25 debugging — shutdown with Close() drops in-flight requests.
//
// On a deploy the orchestrator sends SIGTERM. This server reacts by calling
// srv.Close(), which immediately rips every active connection. An in-flight
// request that was 100ms from finishing gets a broken connection → the
// client sees a 502/EOF. That's the "every deploy causes a blip of 5xx"
// symptom. The correct call is srv.Shutdown(ctx), which drains in-flight
// requests first.
//
// STDLIB ONLY. We start a real server on a random localhost port, launch a
// slow in-flight request, trigger shutdown mid-request, and observe the
// dropped request. Exits promptly (no SIGTERM wait, no hung server).
//
// Run with: go run -race .
package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

func main() {
	mux := http.NewServeMux()
	// A "slow" handler: takes 200ms, like a real request mid-DB-call.
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
		resp, err := client.Get(url) // in-flight request
		if err != nil {
			reqErr = err
			return
		}
		defer resp.Body.Close()
		status = resp.StatusCode
	}()

	// Let the request get in-flight, then "receive SIGTERM" and shut down.
	time.Sleep(50 * time.Millisecond)

	// BUG: Close() immediately closes all active connections — it does NOT
	// wait for the in-flight /work request to finish.
	_ = srv.Close()

	wg.Wait()

	if reqErr != nil {
		fmt.Printf("in-flight request FAILED: %v\n", reqErr)
		fmt.Println("=> BUG: srv.Close() dropped an in-flight request (502/EOF on every deploy)")
		return
	}
	fmt.Printf("in-flight request completed with status %d\n", status)
}

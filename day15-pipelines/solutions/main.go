// Day 15 Phase-3 capstone — reference solution. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

func fakeCheck(ctx context.Context, url string, latency time.Duration) (string, error) {
	select {
	case <-time.After(latency):
		return "ok", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type job struct {
	url     string
	latency time.Duration
}
type result struct {
	url    string
	status string
}

func checkURLs(parent context.Context, urls map[string]time.Duration, workers int, perCheck time.Duration) map[string]string {
	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ctx, cancel := context.WithTimeout(parent, perCheck)
				status, err := fakeCheck(ctx, j.url, j.latency)
				cancel()
				if err != nil {
					status = "timeout"
				}
				results <- result{j.url, status}
			}
		}()
	}

	go func() {
		for u, lat := range urls {
			jobs <- job{u, lat}
		}
		close(jobs)
	}()
	go func() { wg.Wait(); close(results) }()

	out := make(map[string]string)
	for r := range results {
		out[r.url] = r.status
	}
	return out
}

// ---- pipeline ----
func gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
func filterEven(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 != 0 {
				continue
			}
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func main() {
	fmt.Println("== Part 1: URL checker (perCheck=50ms) ==")
	urls := map[string]time.Duration{
		"fast.com":   10 * time.Millisecond,
		"slow.com":   200 * time.Millisecond, // will time out
		"medium.com": 20 * time.Millisecond,
		"stuck.com":  500 * time.Millisecond, // will time out
	}
	res := checkURLs(context.Background(), urls, 3, 50*time.Millisecond)
	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-12s %s\n", k, res[k])
	}

	fmt.Println("== Part 2: pipeline, take first 2 then cancel ==")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := filterEven(ctx, square(ctx, gen(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))
	count := 0
	for v := range out {
		fmt.Println("  got", v)
		count++
		if count == 2 {
			cancel()
			break
		}
	}
	fmt.Println("  done (upstream cancelled cleanly)")
}

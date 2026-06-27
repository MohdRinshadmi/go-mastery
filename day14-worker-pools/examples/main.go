// Day 14 walkthrough — worker pool, fan-in, errgroup.
// Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ---- Worker pool: N workers square numbers ------------------------------
func square(n int) int { return n * n }

func runPool(nums []int, workers int) []int {
	jobs := make(chan int)
	results := make(chan int)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs { // pull until jobs closed
				results <- square(j)
			}
		}()
	}

	// producer: feed jobs then close
	go func() {
		for _, n := range nums {
			jobs <- n
		}
		close(jobs)
	}()

	// closer: close results once all workers finish
	go func() { wg.Wait(); close(results) }()

	var out []int
	for r := range results { // drain until closed
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

// ---- Fan-in: merge channels --------------------------------------------
func fanIn(chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

func gen(vals ...int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for _, v := range vals {
			c <- v
		}
	}()
	return c
}

// ---- errgroup: parallel work with error + cancellation ------------------
func fetchAll(ctx context.Context, ids []int) (map[int]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(3) // bounded concurrency

	var mu sync.Mutex
	out := make(map[int]string)

	for _, id := range ids {
		id := id
		g.Go(func() error {
			select {
			case <-time.After(10 * time.Millisecond): // simulate I/O
			case <-ctx.Done():
				return ctx.Err()
			}
			mu.Lock()
			out[id] = fmt.Sprintf("item-%d", id)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	fmt.Println("== Worker pool (squares, 4 workers) ==")
	fmt.Println("  ", runPool([]int{1, 2, 3, 4, 5, 6, 7, 8}, 4))

	fmt.Println("== Fan-in (merge 3 generators) ==")
	merged := fanIn(gen(1, 2), gen(3, 4), gen(5, 6))
	var got []int
	for v := range merged {
		got = append(got, v)
	}
	sort.Ints(got)
	fmt.Println("  ", got)

	fmt.Println("== errgroup parallel fetch (limit 3) ==")
	res, err := fetchAll(context.Background(), []int{1, 2, 3, 4, 5})
	if err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Printf("  fetched %d items, e.g. id3=%q\n", len(res), res[3])
	}
}

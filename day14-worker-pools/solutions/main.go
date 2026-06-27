// Day 14 — reference solutions. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func squarePool(nums []int, workers int) []int {
	jobs := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results <- j * j
			}
		}()
	}
	go func() {
		for _, n := range nums {
			jobs <- n
		}
		close(jobs)
	}()
	go func() { wg.Wait(); close(results) }()

	var out []int
	for r := range results {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

func merge(chans ...<-chan int) <-chan int {
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

func checkAll(ctx context.Context, items []int) (map[int]bool, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	var mu sync.Mutex
	out := make(map[int]bool)
	for _, it := range items {
		it := it
		g.Go(func() error {
			select {
			case <-time.After(5 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			mu.Lock()
			out[it] = it%2 == 0 // fake "check passed if even"
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
	fmt.Println("== Exercise 1 ==")
	fmt.Println("  ", squarePool([]int{1, 2, 3, 4, 5, 6, 7, 8}, 4))

	fmt.Println("== Exercise 2 ==")
	var got []int
	for v := range merge(gen(1, 2), gen(3, 4), gen(5)) {
		got = append(got, v)
	}
	sort.Ints(got)
	fmt.Println("  ", got)

	fmt.Println("== Challenge ==")
	res, err := checkAll(context.Background(), []int{1, 2, 3, 4})
	fmt.Printf("  results=%v err=%v\n", res, err)
}

// Day 15 walkthrough — cancellable pipeline. Run: go run -race main.go
package main

import (
	"context"
	"fmt"
)

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
	fmt.Println("== Full pipeline: gen -> square -> filterEven ==")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for v := range filterEven(ctx, square(ctx, gen(ctx, 1, 2, 3, 4, 5, 6))) {
		fmt.Println("  ", v) // 4 16 36
	}

	fmt.Println("== Early cancel: take first 2, then cancel (no leak) ==")
	ctx2, cancel2 := context.WithCancel(context.Background())
	out := square(ctx2, gen(ctx2, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	count := 0
	for v := range out {
		fmt.Println("  got", v)
		count++
		if count == 2 {
			cancel2() // tells every upstream stage to stop -> goroutines exit
			break
		}
	}
	fmt.Println("  cancelled after 2; upstream stages tear down cleanly")
}

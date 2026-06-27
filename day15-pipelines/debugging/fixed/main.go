// Day 15 debugging — FIXED.
//
// Same gen -> square pipeline, but every stage propagates cancellation.
//
// Fix: thread a context.Context through each stage and `select` on ctx.Done()
// for every send. The consumer `defer cancel()`s, so when it takes the first few
// results and returns, cancel() tears down the whole pipeline: each stage's send
// loses the race to ctx.Done(), the goroutine returns, defer close(out) runs, and
// the close cascades. No goroutine leaks.
package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func gen(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; ; i++ {
			select {
			case out <- i:
			case <-ctx.Done(): // downstream gave up -> exit, no leak
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

func main() {
	base := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // tears down the whole pipeline when main returns

	results := square(ctx, gen(ctx))
	taken := 0
	for v := range results {
		fmt.Println("got:", v)
		taken++
		if taken == 3 {
			cancel() // tell upstream to stop, then leave
			break
		}
	}

	// Give the cancelled stages a moment to unwind.
	time.Sleep(100 * time.Millisecond)

	leaked := runtime.NumGoroutine() - base
	fmt.Printf("after taking %d results: %d goroutines still alive\n", taken, leaked)
	if leaked <= 0 {
		fmt.Println("no leak: cancel() propagated through every stage, all goroutines exited")
	}
}

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	const workers = 1000

	var wg sync.WaitGroup
	var counter atomic.Int64 // FIX: atomic counter, no lost updates

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1) // single indivisible read-add-write
		}()
	}

	wg.Wait()

	fmt.Printf("counter = %d (want %d)\n", counter.Load(), workers)
}

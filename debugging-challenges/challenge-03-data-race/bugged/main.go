package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers = 1000

	var wg sync.WaitGroup
	counter := 0 // shared, unsynchronized

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// BUG: counter++ is read-modify-write with no synchronization.
			// Concurrent goroutines lose increments. `go run -race .` flags this.
			counter++
		}()
	}

	wg.Wait()

	fmt.Printf("counter = %d (want %d)\n", counter, workers)
}

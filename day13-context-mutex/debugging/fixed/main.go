// Day 13 debugging — FIXED.
//
// Same metrics store, made concurrency-safe.
//
// Fix: guard the map with a sync.RWMutex that lives next to the data and is
// encapsulated inside the methods. Writes take Lock; the read-only Total takes
// RLock (many concurrent readers allowed). Now every access is synchronized:
// counts are exact and `go run -race .` is clean.
//
// We also thread a context so the workload is cancellable, with `defer cancel()`
// to release its resources — the Day 13 discipline for any cancellable work.
package main

import (
	"context"
	"fmt"
	"sync"
)

// Store is safe for concurrent use: the mutex guards the map and is unexported,
// living next to the data it protects.
type Store struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewStore() *Store {
	return &Store{m: make(map[string]int)}
}

func (s *Store) Inc(key string) {
	s.mu.Lock()
	defer s.mu.Unlock() // defer so the lock releases even on panic
	s.m[key]++
}

func (s *Store) Total() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t := 0
	for _, v := range s.m {
		t += v
	}
	return t
}

func main() {
	const goroutines = 50
	const perG = 200

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ALWAYS cancel to release context resources

	s := NewStore()
	keys := []string{"get", "set", "del", "scan"}

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				select {
				case <-ctx.Done():
					return // cancellable
				default:
					s.Inc(keys[i%len(keys)])
				}
			}
		}(g)
	}
	wg.Wait()

	want := goroutines * perG
	got := s.Total()
	fmt.Printf("recorded %d hits, store totals %d\n", want, got)
	if got == want {
		fmt.Println("all updates accounted for — no race, no lost increments")
	}
}

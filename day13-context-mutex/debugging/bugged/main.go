// Day 13 debugging — BUGGED.
//
// Scenario: an in-memory "metrics" store. Many goroutines record hits for
// different keys concurrently. We expect the final per-key counts to sum to the
// number of recordHit calls.
//
// Bug: the underlying map is read and written from many goroutines with NO
// synchronization. This is a textbook data race. Two failure modes:
//   1. `go run -race .` prints WARNING: DATA RACE every time (the read-modify-
//      write `m[key]++` and concurrent reads are unsynchronized).
//   2. Without -race, concurrent map writes can trigger the runtime's built-in
//      detector: `fatal error: concurrent map writes` — a hard crash.
//
// Counts also come out wrong (lost updates) because `m[key]++` is load-add-store.
package main

import (
	"fmt"
	"sync"
)

// Store is NOT safe for concurrent use — that's the bug.
type Store struct {
	m map[string]int
}

func NewStore() *Store {
	return &Store{m: make(map[string]int)}
}

func (s *Store) Inc(key string) {
	s.m[key]++ // BUG: unsynchronized read-modify-write on a shared map
}

func (s *Store) Total() int {
	t := 0
	for _, v := range s.m { // BUG: unsynchronized read concurrent with writes
		t += v
	}
	return t
}

func main() {
	const goroutines = 50
	const perG = 200

	s := NewStore()
	keys := []string{"get", "set", "del", "scan"}

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				s.Inc(keys[i%len(keys)]) // concurrent map writes — racy
			}
		}(g)
	}
	wg.Wait()

	want := goroutines * perG
	got := s.Total()
	fmt.Printf("recorded %d hits, store totals %d\n", want, got)
	if got != want {
		fmt.Printf("BUG: %d updates lost to the data race\n", want-got)
	}
}

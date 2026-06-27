// Day 10 Phase-2 capstone (solution) — storage abstraction.
// go test -bench=. -benchmem ./...
package store

import "sync"

// Store is the abstraction. Two implementations satisfy it.
type Store interface {
	Set(key, value string)
	Get(key string) (string, bool)
}

// MapStore: O(1) average lookups, guarded by a mutex (Phase 3 preview).
type MapStore struct {
	mu sync.RWMutex
	m  map[string]string
}

func NewMapStore() *MapStore { return &MapStore{m: make(map[string]string)} }

func (s *MapStore) Set(key, value string) {
	s.mu.Lock()
	s.m[key] = value
	s.mu.Unlock()
}

func (s *MapStore) Get(key string) (string, bool) {
	s.mu.RLock()
	v, ok := s.m[key]
	s.mu.RUnlock()
	return v, ok
}

// SliceStore: O(n) linear scan — fine for tiny N, degrades badly as N grows.
type kv struct{ k, v string }

type SliceStore struct {
	items []kv
}

func NewSliceStore() *SliceStore { return &SliceStore{} }

func (s *SliceStore) Set(key, value string) {
	for i := range s.items {
		if s.items[i].k == key {
			s.items[i].v = value
			return
		}
	}
	s.items = append(s.items, kv{key, value})
}

func (s *SliceStore) Get(key string) (string, bool) {
	for i := range s.items {
		if s.items[i].k == key {
			return s.items[i].v, true
		}
	}
	return "", false
}

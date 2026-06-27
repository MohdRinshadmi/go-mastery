// Day 10 Phase-2 capstone — storage abstraction. Fill the TODOs.
// Goal: two Store implementations + one shared test suite + comparative
// benchmarks. Run: go test -bench=. -benchmem ./...
package store

// TODO: define the Store interface with:
//   Set(key, value string)
//   Get(key string) (string, bool)
type Store interface {
	// TODO
}

// MapStore — back it with a map[string]string.
type MapStore struct {
	// TODO
}

func NewMapStore() *MapStore { return &MapStore{ /* TODO */ } }

func (s *MapStore) Set(key, value string) {
	// TODO
}

func (s *MapStore) Get(key string) (string, bool) {
	// TODO
	return "", false
}

// SliceStore — back it with a slice of key/value pairs (linear scan).
type SliceStore struct {
	// TODO
}

func NewSliceStore() *SliceStore { return &SliceStore{} }

func (s *SliceStore) Set(key, value string) {
	// TODO (remember: overwrite if key exists)
}

func (s *SliceStore) Get(key string) (string, bool) {
	// TODO
	return "", false
}

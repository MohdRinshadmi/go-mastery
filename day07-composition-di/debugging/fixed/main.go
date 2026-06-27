package main

import "fmt"

// Store is the contract. GetAll is expected to fan out to Get for every id.
type Store interface {
	Get(id string) (string, error)
	GetAll(ids []string) ([]string, error)
}

// baseStore is the real implementation.
type baseStore struct {
	data map[string]string
}

func newBaseStore() *baseStore {
	return &baseStore{data: map[string]string{
		"1": "alice",
		"2": "bob",
		"3": "carol",
	}}
}

func (b *baseStore) Get(id string) (string, error) {
	v, ok := b.data[id]
	if !ok {
		return "", fmt.Errorf("not found: %s", id)
	}
	return v, nil
}

func (b *baseStore) GetAll(ids []string) ([]string, error) {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		v, err := b.Get(id)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// LoggingStore decorates a Store by embedding the interface and overriding Get.
type LoggingStore struct {
	Store // embedded interface: promotes Get and GetAll
}

func newLoggingStore(s Store) *LoggingStore {
	return &LoggingStore{Store: s}
}

func (l *LoggingStore) Get(id string) (string, error) {
	fmt.Printf("[log] Get(%q)\n", id)
	return l.Store.Get(id)
}

// THE FIX: implement GetAll on the decorator too, and route every id through
// l.Get — the decorator's own (logged) Get. Go has no virtual dispatch, so the
// only way to make GetAll honor the override is to write the fan-out here, on
// the receiver that actually owns the override.
func (l *LoggingStore) GetAll(ids []string) ([]string, error) {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		v, err := l.Get(id) // l.Get => LoggingStore.Get (logged)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func main() {
	var store Store = newLoggingStore(newBaseStore())

	fmt.Println("=== direct Get (override fires) ===")
	v, _ := store.Get("1")
	fmt.Println("got:", v)

	fmt.Println("\n=== GetAll (override now fires for every id) ===")
	all, _ := store.GetAll([]string{"2", "3"})
	fmt.Println("got:", all)

	fmt.Println("\nExpected 3 [log] lines total; fixed run prints all 3.")
}

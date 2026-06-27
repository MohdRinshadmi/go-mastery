package main

import "fmt"

// Store is the contract. GetAll is expected to fan out to Get for every id.
type Store interface {
	Get(id string) (string, error)
	GetAll(ids []string) ([]string, error)
}

// baseStore is the real implementation. Note GetAll loops and calls Get.
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

// GetAll fans out to b.Get. Because the receiver here is *baseStore, the
// "b.Get" call is statically bound to baseStore.Get — there is no virtual
// dispatch back up to any decorator that wrapped this store.
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

// LoggingStore decorates a Store by embedding the interface and overriding Get
// to add a log line. The intent: EVERY Get — including the ones GetAll makes —
// should be logged.
type LoggingStore struct {
	Store // embedded interface: promotes Get and GetAll
}

func newLoggingStore(s Store) *LoggingStore {
	return &LoggingStore{Store: s}
}

// Override Get to add logging, then delegate to the wrapped store.
func (l *LoggingStore) Get(id string) (string, error) {
	fmt.Printf("[log] Get(%q)\n", id)
	return l.Store.Get(id)
}

// NOTE: there is no GetAll method here. GetAll is promoted from the embedded
// Store interface, so it runs entirely inside baseStore — whose Get is the
// plain, unlogged one. The override above is never reached on the GetAll path.

func main() {
	var store Store = newLoggingStore(newBaseStore())

	fmt.Println("=== direct Get (override DOES fire) ===")
	v, _ := store.Get("1")
	fmt.Println("got:", v)

	fmt.Println("\n=== GetAll (override SILENTLY skipped — BUG) ===")
	all, _ := store.GetAll([]string{"2", "3"})
	fmt.Println("got:", all)

	fmt.Println("\nExpected 3 [log] lines total; bugged run prints only 1.")
}

// Final exam — Section B coding starters. Implement these, then write
// final_test.go and run `go test -race ./...`. Reference: ../solutions/.
package finalexam

import (
	"context"
	"time"
)

// B1 — generic GroupBy
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	// TODO
	return nil
}

// B2 — concurrency-safe bounded cache; evict oldest-inserted on overflow.
type LRU struct {
	// TODO
}

func NewLRU(max int) *LRU { return &LRU{} }
func (c *LRU) Get(k string) (string, bool) {
	// TODO
	return "", false
}
func (c *LRU) Set(k, v string) {
	// TODO
}

// B3 — bounded URL checker with per-check timeout via context.
func CheckURLs(parent context.Context, urls []string, workers int, perCheck time.Duration,
	check func(ctx context.Context, url string) error) map[string]error {
	// TODO
	return nil
}

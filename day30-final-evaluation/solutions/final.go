// Final exam coding solutions (B1/B2/B3). Tests in final_test.go.
package finalexam

import (
	"context"
	"sync"
	"time"
)

// B1 — generic GroupBy
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, it := range items {
		k := key(it)
		out[k] = append(out[k], it)
	}
	return out
}

// B2 — concurrency-safe bounded cache with oldest-inserted eviction
type LRU struct {
	mu    sync.Mutex
	max   int
	store map[string]string
	order []string
}

func NewLRU(max int) *LRU {
	return &LRU{max: max, store: make(map[string]string)}
}
func (c *LRU) Get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.store[k]
	return v, ok
}
func (c *LRU) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[k]; !ok {
		if len(c.order) >= c.max {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.store, oldest)
		}
		c.order = append(c.order, k)
	}
	c.store[k] = v
}

// B3 — bounded URL checker with per-check timeout
func CheckURLs(parent context.Context, urls []string, workers int, perCheck time.Duration,
	check func(ctx context.Context, url string) error) map[string]error {

	jobs := make(chan string)
	type res struct {
		url string
		err error
	}
	results := make(chan res)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range jobs {
				ctx, cancel := context.WithTimeout(parent, perCheck)
				err := check(ctx, u)
				cancel()
				results <- res{u, err}
			}
		}()
	}
	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()
	go func() { wg.Wait(); close(results) }()

	out := make(map[string]error)
	for r := range results {
		out[r.url] = r.err
	}
	return out
}

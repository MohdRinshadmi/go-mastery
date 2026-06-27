// Day 28 — YOUR exercises. Run: go run .
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type DB struct{ calls atomic.Int64 }

func (d *DB) Get(_ context.Context, key string) (string, error) {
	d.calls.Add(1)
	time.Sleep(50 * time.Millisecond)
	return "value-of-" + key, nil
}

// =====================================================================
// TASK 1 — cache-aside Cache with per-entry TTL.
//   Get(ctx, key): hit -> return; miss -> db.Get, store with TTL, return.
//
// TASK 2 — stampede protection with singleflight so 50 concurrent misses
//   for the same key cause exactly ONE db.Get.
//   (go get golang.org/x/sync/singleflight && go mod tidy)
//
// CHALLENGE — cache negative lookups: have DB.Get sometimes return a
//   "not found" sentinel; cache that result (short TTL) so repeated misses
//   for a non-existent key don't keep hitting the DB (cache penetration).
// =====================================================================

type Cache struct {
	// TODO: db, ttl, store map[string]entry, singleflight.Group
}

func NewCache(db *DB, ttl time.Duration) *Cache {
	// TODO
	return &Cache{}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	// TODO
	return "", nil
}

func main() {
	db := &DB{}
	_ = NewCache(db, time.Minute)
	// TODO: fire 50 concurrent gets for one key, then print db.calls (want 1)
	fmt.Println("TODO: implement cache-aside + singleflight; see ../solutions")
}

//go:build ignore

// REFERENCE ONLY — real Redis (go-redis) cache-aside + distributed lock.
// Build-tagged `ignore`. To run:
//   docker run -p 6379:6379 redis:7-alpine
//   remove the build tag, go get github.com/redis/go-redis/v9, go run .
package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RedisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) *RedisCache {
	return &RedisCache{rdb: redis.NewClient(&redis.Options{Addr: addr}), ttl: ttl}
}

// cache-aside read against Redis
func (c *RedisCache) GetProduct(ctx context.Context, id string, load func(context.Context, string) (Product, error)) (Product, error) {
	key := "product:" + id
	if b, err := c.rdb.Get(ctx, key).Bytes(); err == nil {
		var p Product
		if json.Unmarshal(b, &p) == nil {
			return p, nil // hit
		}
	} else if !errors.Is(err, redis.Nil) {
		// real error (not a miss) — log + fall through
	}
	p, err := load(ctx, id) // miss -> source of truth
	if err != nil {
		return Product{}, err
	}
	if b, err := json.Marshal(p); err == nil {
		c.rdb.Set(ctx, key, b, c.ttl) // populate with TTL
	}
	return p, nil
}

func (c *RedisCache) Invalidate(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, "product:"+id).Err()
}

// distributed lock via SETNX (lease) — basic stampede/critical-section guard.
// Production: use Redlock or a proper lock library; handle lock expiry/renewal.
func (c *RedisCache) tryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, "lock:"+key, "1", ttl).Result()
}

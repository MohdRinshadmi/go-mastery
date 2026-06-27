// Day 19 walkthrough — repository pattern, runs offline with in-memory repo.
// Run: go run .
//
// The real Postgres + Redis implementations live in postgres_reference.go
// (build-tagged `ignore` so they don't require the drivers to compile here).
// See docker-compose.yml to run the real stack.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ---- Domain ----
type User struct {
	ID    string
	Email string
	Name  string
}

var ErrUserNotFound = errors.New("user not found")

// ---- The interface the SERVICE depends on (owned here, not by postgres) --
type UserRepository interface {
	Create(ctx context.Context, u User) error
	GetByID(ctx context.Context, id string) (User, error)
}

// ---- In-memory implementation (used in tests + this offline demo) --------
type InMemoryUserRepo struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{users: make(map[string]User)}
}
func (r *InMemoryUserRepo) Create(_ context.Context, u User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}
func (r *InMemoryUserRepo) GetByID(_ context.Context, id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

// ---- CachingUserRepo: wraps ANY UserRepository (decorator via composition)
type cacheEntry struct {
	u   User
	exp time.Time
}
type CachingUserRepo struct {
	inner UserRepository
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]cacheEntry
	Hits  int
	Miss  int
}

func NewCachingUserRepo(inner UserRepository, ttl time.Duration) *CachingUserRepo {
	return &CachingUserRepo{inner: inner, ttl: ttl, cache: make(map[string]cacheEntry)}
}
func (r *CachingUserRepo) Create(ctx context.Context, u User) error {
	if err := r.inner.Create(ctx, u); err != nil {
		return err
	}
	r.mu.Lock()
	delete(r.cache, u.ID) // invalidate on write
	r.mu.Unlock()
	return nil
}
func (r *CachingUserRepo) GetByID(ctx context.Context, id string) (User, error) {
	r.mu.Lock()
	if e, ok := r.cache[id]; ok && time.Now().Before(e.exp) {
		r.Hits++
		r.mu.Unlock()
		return e.u, nil
	}
	r.Miss++
	r.mu.Unlock()

	u, err := r.inner.GetByID(ctx, id) // cache miss -> hit the inner repo
	if err != nil {
		return User{}, err
	}
	r.mu.Lock()
	r.cache[id] = cacheEntry{u: u, exp: time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return u, nil
}

func main() {
	ctx := context.Background()

	// Compose: cache in front of in-memory repo. Same interface throughout.
	var repo UserRepository = NewCachingUserRepo(NewInMemoryUserRepo(), time.Minute)
	cached := repo.(*CachingUserRepo)

	_ = repo.Create(ctx, User{ID: "u1", Email: "ada@x.com", Name: "Ada"})

	fmt.Println("== Repository pattern demo ==")
	for i := 0; i < 3; i++ {
		u, err := repo.GetByID(ctx, "u1")
		fmt.Printf("  get u1 -> %s (err=%v)\n", u.Name, err)
	}
	fmt.Printf("  cache hits=%d misses=%d (1 miss to load, then hits)\n", cached.Hits, cached.Miss)

	_, err := repo.GetByID(ctx, "missing")
	fmt.Printf("  get missing -> domain error: %v\n", err)
}

// Day 19 — reference solution. Run: go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type User struct {
	ID    string
	Email string
	Name  string
}

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, u User) error
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

// ---- In-memory ----
type InMemoryUserRepo struct {
	mu    sync.RWMutex
	byID  map[string]User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{byID: make(map[string]User)}
}
func (r *InMemoryUserRepo) Create(_ context.Context, u User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[u.ID] = u
	return nil
}
func (r *InMemoryUserRepo) GetByID(_ context.Context, id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}
func (r *InMemoryUserRepo) GetByEmail(_ context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.byID {
		if u.Email == email {
			return u, nil
		}
	}
	return User{}, ErrUserNotFound
}

// ---- Caching decorator ----
type entry struct {
	u   User
	exp time.Time
}
type CachingUserRepo struct {
	inner UserRepository
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]entry
	Hits  int
	Miss  int
}

func NewCachingUserRepo(inner UserRepository, ttl time.Duration) *CachingUserRepo {
	return &CachingUserRepo{inner: inner, ttl: ttl, cache: make(map[string]entry)}
}
func (r *CachingUserRepo) Create(ctx context.Context, u User) error {
	if err := r.inner.Create(ctx, u); err != nil {
		return err
	}
	r.mu.Lock()
	delete(r.cache, u.ID)
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
	u, err := r.inner.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	r.mu.Lock()
	r.cache[id] = entry{u: u, exp: time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return u, nil
}
func (r *CachingUserRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	return r.inner.GetByEmail(ctx, email)
}

func main() {
	ctx := context.Background()
	repo := NewCachingUserRepo(NewInMemoryUserRepo(), time.Minute)

	_ = repo.Create(ctx, User{ID: "u1", Email: "ada@x.com", Name: "Ada"})

	fmt.Println("== Repository demo ==")
	for i := 0; i < 3; i++ {
		u, _ := repo.GetByID(ctx, "u1")
		fmt.Printf("  get u1 -> %s\n", u.Name)
	}
	fmt.Printf("  hits=%d misses=%d\n", repo.Hits, repo.Miss)

	if _, err := repo.GetByID(ctx, "nope"); errors.Is(err, ErrUserNotFound) {
		fmt.Println("  missing id -> ErrUserNotFound (domain error)")
	}
	u, _ := repo.GetByEmail(ctx, "ada@x.com")
	fmt.Printf("  by email -> %s\n", u.Name)
}

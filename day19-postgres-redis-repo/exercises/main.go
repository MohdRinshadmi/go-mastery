// Day 19 — YOUR exercises. Run: go run .
package main

import (
	"context"
	"errors"
	"fmt"
)

type User struct {
	ID    string
	Email string
	Name  string
}

var ErrUserNotFound = errors.New("user not found")

// The interface your service depends on.
type UserRepository interface {
	Create(ctx context.Context, u User) error
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

// =====================================================================
// EXERCISE 1 — InMemoryUserRepo
// Implement all three methods. Guard with a sync.RWMutex. GetByID/GetByEmail
// return ErrUserNotFound when absent. (GetByEmail: scan the map.)
// =====================================================================

type InMemoryUserRepo struct {
	// TODO
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	// TODO
	return &InMemoryUserRepo{}
}

// TODO: Create / GetByID / GetByEmail

// =====================================================================
// CHALLENGE — CachingUserRepo (decorator)
// Wrap ANY UserRepository with a small in-memory TTL cache for GetByID.
// On Create, invalidate the cached id. Track Hits/Misses. Same interface!
// (GetByEmail can just delegate to inner for now.)
// =====================================================================

type CachingUserRepo struct {
	// TODO: inner UserRepository, ttl, cache map, hit/miss counters
}

func main() {
	ctx := context.Background()
	fmt.Println("== Repository demo ==")
	// TODO: create an InMemoryUserRepo (optionally wrapped in CachingUserRepo),
	// Create a user, GetByID a few times, GetByID a missing id, print results.
	_ = ctx
}

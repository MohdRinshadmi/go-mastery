// Day 19 debugging — FIXED.
//
// The repository returns a COPY of the stored user, so a caller mutating the
// returned value can never reach back into the repository's internal map.
//
// No database, no pgx, no redis — stdlib only, so it builds and runs offline.
package main

import (
	"context"
	"errors"
	"fmt"
)

// ErrUserNotFound is the domain sentinel a repository returns for a missing
// row (the in-memory stand-in for mapping pgx.ErrNoRows at the boundary).
var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID    string
	Email string
	Name  string
}

// UserRepository is the interface the service layer owns and depends on.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
}

// InMemoryUserRepo is a fake repo backed by a map of pointers.
type InMemoryUserRepo struct {
	users map[string]*User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users: map[string]*User{
			"u1": {ID: "u1", Email: "alice@example.com", Name: "Alice"},
		},
	}
}

// GetByID looks up a user by id.
//
// FIX: dereference the stored pointer to make a value copy, then return the
// address of that fresh copy. The caller's *User is its own object; mutating
// it cannot touch the repository's stored User.
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	cp := *u   // copy the struct value
	return &cp, nil
}

func main() {
	ctx := context.Background()
	var repo UserRepository = NewInMemoryUserRepo()

	// First fetch.
	u1, err := repo.GetByID(ctx, "u1")
	if err != nil {
		fmt.Println("unexpected error:", err)
		return
	}
	fmt.Printf("fetch #1: %s -> %q\n", u1.ID, u1.Name)

	// A caller mutates the returned user.
	u1.Name = "HACKED"

	// Second, independent fetch of the SAME id from the repo.
	u2, err := repo.GetByID(ctx, "u1")
	if err != nil {
		fmt.Println("unexpected error:", err)
		return
	}
	fmt.Printf("fetch #2: %s -> %q\n", u2.ID, u2.Name)

	if u2.Name == "HACKED" {
		fmt.Println("CORRUPTED: caller's mutation leaked into the repository's stored data")
	} else {
		fmt.Println("OK: stored data is intact")
	}

	// Bonus: a missing id returns the domain sentinel, distinguishable with errors.Is.
	if _, err := repo.GetByID(ctx, "nope"); errors.Is(err, ErrUserNotFound) {
		fmt.Println("missing id -> ErrUserNotFound (mapped at the boundary)")
	}
}

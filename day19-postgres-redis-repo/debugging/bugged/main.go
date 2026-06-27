// Day 19 debugging — BUGGED.
//
// An in-memory UserRepository stores users in a map[string]*User and its
// GetByID returns the STORED pointer directly. A caller that mutates the
// returned *User silently corrupts the repository's own copy — the map now
// points at the caller's edit. This simulates the real-world repo trap of
// handing out internal references instead of copies.
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
// BUG: it returns the stored *User straight out of the map. Every caller gets
// the SAME pointer the repository keeps internally, so any mutation the caller
// makes is reflected in the repository's stored data — aliasing corruption.
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil // BUG: shared internal pointer, not a copy
}

func main() {
	ctx := context.Background()
	var repo UserRepository = NewInMemoryUserRepo()

	// First fetch — looks fine.
	u1, err := repo.GetByID(ctx, "u1")
	if err != nil {
		fmt.Println("unexpected error:", err)
		return
	}
	fmt.Printf("fetch #1: %s -> %q\n", u1.ID, u1.Name)

	// A caller mutates the returned user (e.g. building a redacted view).
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
}

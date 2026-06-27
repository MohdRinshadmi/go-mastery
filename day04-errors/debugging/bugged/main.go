package main

import (
	"errors"
	"fmt"
)

// ErrNotFound is the sentinel returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// lookup simulates a data-access call. For unknown IDs it returns ErrNotFound,
// but WRAPPED with context (as every good Go layer does).
func lookup(id int) (string, error) {
	users := map[int]string{1: "Ada", 2: "Linus"}
	name, ok := users[id]
	if !ok {
		// Wrap the sentinel with %w to add context.
		return "", fmt.Errorf("lookup id=%d: %w", id, ErrNotFound)
	}
	return name, nil
}

// greet returns a friendly message, treating "not found" as a soft, expected
// case (a default greeting) and anything else as a hard error.
func greet(id int) (string, error) {
	name, err := lookup(id)
	if err != nil {
		// BUG: == only matches the exact top-level error value. Because lookup
		// WRAPPED the sentinel, err is *fmt.wrapError, not ErrNotFound itself,
		// so this comparison is always false.
		if err == ErrNotFound {
			return "Hello, stranger!", nil
		}
		return "", err
	}
	return "Hello, " + name + "!", nil
}

func main() {
	for _, id := range []int{1, 99} {
		msg, err := greet(id)
		if err != nil {
			// id=99 should have been handled softly, but it leaks out here.
			fmt.Printf("id=%d ERROR: %v\n", id, err)
			continue
		}
		fmt.Printf("id=%d -> %s\n", id, msg)
	}
}

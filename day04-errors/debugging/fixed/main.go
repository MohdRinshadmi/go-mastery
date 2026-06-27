package main

import (
	"errors"
	"fmt"
)

// ErrNotFound is the sentinel returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// lookup simulates a data-access call. For unknown IDs it returns ErrNotFound
// wrapped with context.
func lookup(id int) (string, error) {
	users := map[int]string{1: "Ada", 2: "Linus"}
	name, ok := users[id]
	if !ok {
		return "", fmt.Errorf("lookup id=%d: %w", id, ErrNotFound)
	}
	return name, nil
}

// greet treats "not found" as a soft case and anything else as a hard error.
func greet(id int) (string, error) {
	name, err := lookup(id)
	if err != nil {
		// FIX: errors.Is walks the whole wrap chain, so it matches the sentinel
		// even though it was wrapped with %w.
		if errors.Is(err, ErrNotFound) {
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
			fmt.Printf("id=%d ERROR: %v\n", id, err)
			continue
		}
		fmt.Printf("id=%d -> %s\n", id, msg)
	}
}

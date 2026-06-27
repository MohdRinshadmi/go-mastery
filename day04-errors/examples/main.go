// Day 04 walkthrough — error handling the Go way.
// Run: go run main.go
package main

import (
	"errors"
	"fmt"
)

// ---- 3. Sentinel errors -------------------------------------------------
var ErrNotFound = errors.New("not found")

type User struct {
	ID   string
	Name string
}

var db = map[string]User{"42": {ID: "42", Name: "Ada"}}

func getUser(id string) (User, error) {
	u, ok := db[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

// ---- 4. Wrapping with %w ------------------------------------------------
func loadProfile(id string) (User, error) {
	u, err := getUser(id)
	if err != nil {
		// add context, keep the cause inspectable
		return User{}, fmt.Errorf("loadProfile %s: %w", id, err)
	}
	return u, nil
}

// ---- 5. Custom error type carrying data ---------------------------------
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Msg)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Msg: "must be non-negative"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Msg: "unrealistically large"}
	}
	return nil
}

// ---- 6. panic / recover -------------------------------------------------
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil // b==0 panics; recover converts it to an error
}

func main() {
	fmt.Println("== Sentinel errors (errors.Is) ==")
	if _, err := loadProfile("999"); err != nil {
		fmt.Println("  got:", err)
		// errors.Is walks the wrap chain to find the sentinel
		if errors.Is(err, ErrNotFound) {
			fmt.Println("  -> recognized as ErrNotFound (would map to 404)")
		}
	}
	if u, err := loadProfile("42"); err == nil {
		fmt.Printf("  loaded: %s\n", u.Name)
	}

	fmt.Println("== Custom error types (errors.As) ==")
	err := validateAge(-3)
	fmt.Println("  got:", err)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("  -> structured: field=%s msg=%s (would map to 400)\n", ve.Field, ve.Msg)
	}

	fmt.Println("== panic/recover at a boundary ==")
	if _, err := safeDivide(10, 0); err != nil {
		fmt.Println("  survived divide-by-zero:", err)
	}
	if r, err := safeDivide(10, 2); err == nil {
		fmt.Println("  10/2 =", r)
	}
}

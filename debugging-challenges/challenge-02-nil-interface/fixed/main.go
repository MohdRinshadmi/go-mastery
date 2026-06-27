package main

import "fmt"

type User struct {
	Name string
}

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Reason
}

// validateUser returns the untyped nil literal on success, so the returned
// error interface is genuinely nil.
func validateUser(u User) error {
	if u.Name == "" {
		return &ValidationError{Reason: "name required"}
	}
	return nil // FIX: untyped nil -> truly nil interface
}

func main() {
	u := User{Name: "Ada"}

	if err := validateUser(u); err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("user is valid")
}

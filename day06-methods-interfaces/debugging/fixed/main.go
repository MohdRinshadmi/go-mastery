package main

import "fmt"

// ValidationError is a concrete error type with a POINTER receiver.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// validate checks a username. It returns an error if invalid, nil if valid.
//
// THE FIX: the function's return type is `error` (an interface). On the
// success path we return the untyped `nil` literal directly, so the caller
// receives a true nil interface (type=nil, value=nil). We never let a
// concrete typed nil pointer get boxed into the interface.
func validate(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Msg: "must not be empty"}
	}
	return nil // untyped nil -> genuine nil interface
}

func main() {
	fmt.Println("=== fixed ===")

	// Failure path — genuinely invalid.
	if err := validate(""); err != nil {
		fmt.Println("invalid input:", err)
	} else {
		fmt.Println("empty name accepted (should not happen)")
	}

	// Success path — name is valid.
	if err := validate("alice"); err != nil {
		fmt.Println("validation failed:", err)
	} else {
		fmt.Println("alice is valid")
	}
}

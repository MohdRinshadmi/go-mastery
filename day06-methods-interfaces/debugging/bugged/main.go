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
// THE BUG: `e` is declared as a CONCRETE pointer type (*ValidationError).
// On the success path it is never assigned, so it stays a nil *ValidationError.
// But when we `return e`, that nil pointer gets boxed into the `error`
// interface — and the interface now holds (type=*ValidationError, value=nil),
// which is NOT a nil interface. The caller's `err != nil` is TRUE even though
// validation succeeded.
func validate(name string) error {
	var e *ValidationError // concrete nil pointer, NOT untyped nil
	if name == "" {
		e = &ValidationError{Field: "name", Msg: "must not be empty"}
	}
	return e // returns a typed nil on the success path — the trap
}

func main() {
	fmt.Println("=== bugged ===")

	// Failure path — genuinely invalid. This one is correct by accident.
	if err := validate(""); err != nil {
		fmt.Println("invalid input:", err)
	} else {
		fmt.Println("empty name accepted (should not happen)")
	}

	// Success path — name is valid, validate() should signal "no error".
	if err := validate("alice"); err != nil {
		// WRONG: we land here even though "alice" is valid.
		fmt.Println("validation failed:", err) // prints: validation failed: <nil>
	} else {
		fmt.Println("alice is valid")
	}
}

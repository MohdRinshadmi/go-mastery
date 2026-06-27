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

// validateUser returns nil when the user is valid... or does it?
func validateUser(u User) error {
	var verr *ValidationError // concrete typed pointer, starts nil

	if u.Name == "" {
		verr = &ValidationError{Reason: "name required"}
	}

	// BUG: returning a *ValidationError. Even when verr == nil, boxing it into
	// the error interface yields a non-nil interface (type half is set).
	return verr
}

func main() {
	u := User{Name: "Ada"} // perfectly valid

	if err := validateUser(u); err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("user is valid")
}

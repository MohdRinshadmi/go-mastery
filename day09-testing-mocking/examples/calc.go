// Day 09 examples — code under test. Run tests: go test -v ./...
package calc

import "errors"

var ErrDivByZero = errors.New("division by zero")

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivByZero
	}
	return a / b, nil
}

// Charger is a NARROW interface — easy to fake in tests.
type Charger interface {
	Charge(amount int) (string, error)
}

// Checkout depends on the interface, not a concrete gateway -> testable.
func Checkout(c Charger, amount int) (string, error) {
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	return c.Charge(amount)
}

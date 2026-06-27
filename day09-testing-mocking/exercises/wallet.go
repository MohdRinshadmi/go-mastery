// Day 09 exercises — code under test. Your job: write wallet_test.go.
package wallet

import "errors"

var ErrInsufficient = errors.New("insufficient funds")

// Withdraw is PURE logic — perfect for a table-driven test.
func Withdraw(balance, amount int) (int, error) {
	if amount <= 0 {
		return balance, errors.New("amount must be positive")
	}
	if amount > balance {
		return balance, ErrInsufficient
	}
	return balance - amount, nil
}

// Charger is the dependency you'll FAKE in your test.
type Charger interface {
	Charge(amount int) (string, error)
}

// Checkout: validate, then charge via the injected gateway.
func Checkout(c Charger, amount int) (string, error) {
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	return c.Charge(amount)
}

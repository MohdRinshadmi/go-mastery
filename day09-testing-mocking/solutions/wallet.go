// Day 09 solution — code under test.
package wallet

import "errors"

var ErrInsufficient = errors.New("insufficient funds")

func Withdraw(balance, amount int) (int, error) {
	if amount <= 0 {
		return balance, errors.New("amount must be positive")
	}
	if amount > balance {
		return balance, ErrInsufficient
	}
	return balance - amount, nil
}

type Charger interface {
	Charge(amount int) (string, error)
}

func Checkout(c Charger, amount int) (string, error) {
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	return c.Charge(amount)
}

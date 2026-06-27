package main

import (
	"errors"
	"fmt"
)

// withdraw subtracts amount from balance. It returns the new balance and an
// error if the account would go negative.
func withdraw(balance, amount int) (int, error) {
	if amount > balance {
		return balance, errors.New("insufficient funds")
	}
	return balance - amount, nil
}

func main() {
	balance := 100

	withdrawals := []int{30, 50, 40} // the third one overdraws

	var err error // outer err: holds the last failure
	for _, amount := range withdrawals {
		// FIX: declare newBalance separately, then ASSIGN to the outer err
		// with `=`, not `:=`. No new err is created, so the outer one is
		// actually updated.
		var newBalance int
		newBalance, err = withdraw(balance, amount)
		if err != nil {
			fmt.Printf("withdraw %d failed: %v\n", amount, err)
			continue
		}
		balance = newBalance
		fmt.Printf("withdrew %d, balance now %d\n", amount, balance)
	}

	if err != nil {
		fmt.Println("RESULT: some withdrawals failed")
	} else {
		fmt.Println("RESULT: all withdrawals succeeded")
	}
}

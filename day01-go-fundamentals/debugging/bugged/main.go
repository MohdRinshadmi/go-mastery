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

	// We process a list of withdrawals. After the loop we report the final
	// balance and whether ANY withdrawal failed.
	withdrawals := []int{30, 50, 40} // the third one overdraws

	var err error // outer err: should hold the last failure
	for _, amount := range withdrawals {
		newBalance, err := withdraw(balance, amount) // BUG: shadows the outer err
		if err != nil {
			fmt.Printf("withdraw %d failed: %v\n", amount, err)
			continue
		}
		balance = newBalance
		fmt.Printf("withdrew %d, balance now %d\n", amount, balance)
	}

	// The outer err is still nil here — the := inside the loop created a new,
	// inner err each iteration, so this "all good" branch always wins.
	if err != nil {
		fmt.Println("RESULT: some withdrawals failed")
	} else {
		fmt.Println("RESULT: all withdrawals succeeded")
	}
}

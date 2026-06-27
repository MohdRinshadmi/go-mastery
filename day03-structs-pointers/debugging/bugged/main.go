package main

import "fmt"

// Account models a bank account with a running balance.
type Account struct {
	Owner   string
	Balance int
}

// Deposit adds amount to the balance.
//
// BUG: this uses a VALUE receiver (c Account), so it mutates a COPY of the
// account. The caller's Account is never changed.
func (a Account) Deposit(amount int) {
	a.Balance += amount
}

// Withdraw subtracts amount from the balance (same value-receiver bug).
func (a Account) Withdraw(amount int) {
	a.Balance -= amount
}

func main() {
	acc := Account{Owner: "Ada", Balance: 100}

	acc.Deposit(50)
	acc.Withdraw(20)

	// We expect 100 + 50 - 20 = 130.
	fmt.Printf("%s's balance: %d\n", acc.Owner, acc.Balance) // prints 100, not 130
}

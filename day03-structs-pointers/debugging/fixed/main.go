package main

import "fmt"

// Account models a bank account with a running balance.
type Account struct {
	Owner   string
	Balance int
}

// Deposit adds amount to the balance.
//
// FIX: use a POINTER receiver (a *Account) so the method mutates the original
// account, not a copy.
func (a *Account) Deposit(amount int) {
	a.Balance += amount
}

// Withdraw subtracts amount from the balance.
func (a *Account) Withdraw(amount int) {
	a.Balance -= amount
}

func main() {
	acc := Account{Owner: "Ada", Balance: 100}

	// acc is addressable, so Go auto-takes its address: (&acc).Deposit(50).
	acc.Deposit(50)
	acc.Withdraw(20)

	fmt.Printf("%s's balance: %d\n", acc.Owner, acc.Balance) // prints 130
}

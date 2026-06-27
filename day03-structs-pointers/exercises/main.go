// Day 03 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// I (your mentor) will review this like a production PR. Write clean code.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Struct modeling + pointer receiver
//
// Define a BankAccount struct with fields:
//   Owner   string
//   Balance float64
//
// Implement:
//   Deposit(amount float64) error  — pointer receiver, reject amount <= 0
//   Withdraw(amount float64) error — pointer receiver, reject amount <= 0
//                                    OR amount > Balance (insufficient funds)
//   String() string                — value receiver, returns a readable summary
//                                    like: "Alice's account: $150.00"
//
// In main: create an account, deposit $200, withdraw $50, try to withdraw $500
// (expect an error), print the account after each operation.
// =====================================================================

type BankAccount struct {
	// TODO: add fields
}

func (a *BankAccount) Deposit(amount float64) error {
	// TODO: implement
	return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
	// TODO: implement
	return nil
}

func (a BankAccount) String() string {
	// TODO: implement
	return ""
}

// =====================================================================
// EXERCISE 2 (beginner) — Embedding + method promotion
//
// Define:
//   type Shape struct { Color string }
//   func (s Shape) Describe() string — returns "a <Color> shape"
//
//   type Circle struct { embeds Shape; Radius float64 }
//   type Rectangle struct { embeds Shape; Width, Height float64 }
//
// Circle gets Area() float64  → π * r²  (use 3.14159 as π)
// Rectangle gets Area() float64 → w * h
//
// In main: create a Circle and Rectangle, call Describe() on each
// (uses the promoted method), call Area() on each, and print results.
// =====================================================================

type Shape struct {
	// TODO
}

func (s Shape) Describe() string {
	// TODO
	return ""
}

type Circle struct {
	// TODO: embed Shape, add Radius
}

func (c Circle) Area() float64 {
	// TODO
	return 0
}

type Rectangle struct {
	// TODO: embed Shape, add Width and Height
}

func (r Rectangle) Area() float64 {
	// TODO
	return 0
}

// =====================================================================
// EXERCISE 3 (beginner) — Type switch + control flow
//
// Write summarize(values []interface{}) that iterates over values and:
//   - For each int: sums them (track a running total)
//   - For each string: collects them into a []string
//   - For anything else: counts it as "unknown"
// Returns (intSum int, strings []string, unknownCount int)
//
// In main: call summarize with a mixed slice and print the results.
// =====================================================================

func summarize(values []interface{}) (intSum int, strings []string, unknownCount int) {
	// TODO: implement using a type switch
	return
}

// =====================================================================
// CHALLENGE (intermediate) — Linked list with pointer receivers
//
// Implement a simple singly-linked list:
//
//   type Node struct { Val int; Next *Node }
//   type LinkedList struct { Head *Node; size int }
//
// Implement (ALL with pointer receivers):
//   Prepend(v int)         — insert at front: O(1)
//   Append(v int)          — insert at back: O(n)
//   Delete(v int) bool     — delete first occurrence of v, return true if found
//   ToSlice() []int        — return all values in order
//   Len() int              — return current size
//
// Gotcha to avoid: when deleting, handle the case where the node to
// delete is the Head separately (pointer-to-pointer pattern).
//
// In main: build a list [1,2,3,4,5], delete 3, prepend 0,
// append 6, then print the final ToSlice() — should be [0,1,2,4,5,6].
// =====================================================================

type Node struct {
	Val  int
	Next *Node
}

type LinkedList struct {
	Head *Node
	size int
}

func (l *LinkedList) Prepend(v int) {
	// TODO
}

func (l *LinkedList) Append(v int) {
	// TODO
}

func (l *LinkedList) Delete(v int) bool {
	// TODO
	return false
}

func (l *LinkedList) ToSlice() []int {
	// TODO
	return nil
}

func (l *LinkedList) Len() int {
	return l.size
}

func main() {
	fmt.Println("== Exercise 1: BankAccount ==")
	// TODO: demonstrate Deposit, Withdraw, error on overdraft

	fmt.Println("== Exercise 2: Embedding ==")
	// TODO: create Circle{Shape{"red"}, 5} and Rectangle, call Describe+Area

	fmt.Println("== Exercise 3: summarize ==")
	mixed := []interface{}{1, "hello", 3.14, 2, "world", true, 3}
	intSum, strs, unknowns := summarize(mixed)
	fmt.Printf("  intSum=%d  strings=%v  unknowns=%d\n", intSum, strs, unknowns)
	// Expected: intSum=6, strings=[hello world], unknowns=2 (3.14 and true)

	fmt.Println("== Challenge: LinkedList ==")
	// TODO: build list, delete, prepend, append, print ToSlice
}

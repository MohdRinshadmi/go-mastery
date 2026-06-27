// Day 03 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"errors"
	"fmt"
)

// ---- Exercise 1: BankAccount -----------------------------------------------

type BankAccount struct {
	Owner   string
	Balance float64
}

func (a *BankAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}
	a.Balance += amount
	return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}
	if amount > a.Balance {
		return fmt.Errorf("insufficient funds: have %.2f, need %.2f", a.Balance, amount)
	}
	a.Balance -= amount
	return nil
}

// Value receiver for String: reading state, no mutation needed.
func (a BankAccount) String() string {
	return fmt.Sprintf("%s's account: $%.2f", a.Owner, a.Balance)
}

// ---- Exercise 2: Embedding -------------------------------------------------

type Shape struct {
	Color string
}

func (s Shape) Describe() string {
	return "a " + s.Color + " shape"
}

type Circle struct {
	Shape
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

type Rectangle struct {
	Shape
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// ---- Exercise 3: summarize -------------------------------------------------

func summarize(values []interface{}) (intSum int, strings []string, unknownCount int) {
	for _, v := range values {
		switch val := v.(type) {
		case int:
			intSum += val
		case string:
			strings = append(strings, val)
		default:
			unknownCount++
		}
	}
	return
}

// ---- Challenge: LinkedList -------------------------------------------------

type Node struct {
	Val  int
	Next *Node
}

type LinkedList struct {
	Head *Node
	size int
}

func (l *LinkedList) Prepend(v int) {
	l.Head = &Node{Val: v, Next: l.Head}
	l.size++
}

func (l *LinkedList) Append(v int) {
	newNode := &Node{Val: v}
	if l.Head == nil {
		l.Head = newNode
		l.size++
		return
	}
	curr := l.Head
	for curr.Next != nil {
		curr = curr.Next
	}
	curr.Next = newNode
	l.size++
}

func (l *LinkedList) Delete(v int) bool {
	if l.Head == nil {
		return false
	}
	// Special case: deleting the head node
	if l.Head.Val == v {
		l.Head = l.Head.Next
		l.size--
		return true
	}
	// Find the node just before the target
	prev := l.Head
	for prev.Next != nil {
		if prev.Next.Val == v {
			prev.Next = prev.Next.Next // unlink the node
			l.size--
			return true
		}
		prev = prev.Next
	}
	return false
}

func (l *LinkedList) ToSlice() []int {
	result := make([]int, 0, l.size)
	curr := l.Head
	for curr != nil {
		result = append(result, curr.Val)
		curr = curr.Next
	}
	return result
}

func (l *LinkedList) Len() int {
	return l.size
}

func main() {
	fmt.Println("== Exercise 1: BankAccount ==")
	acct := &BankAccount{Owner: "Alice"}
	fmt.Println(" ", acct) // $0.00

	if err := acct.Deposit(200); err != nil {
		fmt.Println("  deposit error:", err)
	}
	fmt.Println(" ", acct) // $200.00

	if err := acct.Withdraw(50); err != nil {
		fmt.Println("  withdraw error:", err)
	}
	fmt.Println(" ", acct) // $150.00

	if err := acct.Withdraw(500); err != nil {
		fmt.Println("  expected error:", err) // insufficient funds
	}
	fmt.Println(" ", acct) // still $150.00

	if err := acct.Deposit(-10); err != nil {
		fmt.Println("  expected error:", err) // negative deposit
	}

	fmt.Println("== Exercise 2: Embedding ==")
	c := Circle{Shape: Shape{Color: "red"}, Radius: 5}
	r := Rectangle{Shape: Shape{Color: "blue"}, Width: 4, Height: 6}

	fmt.Printf("  Circle:    %s  area=%.2f\n", c.Describe(), c.Area())
	fmt.Printf("  Rectangle: %s  area=%.2f\n", r.Describe(), r.Area())

	fmt.Println("== Exercise 3: summarize ==")
	mixed := []interface{}{1, "hello", 3.14, 2, "world", true, 3}
	intSum, strs, unknowns := summarize(mixed)
	fmt.Printf("  intSum=%d  strings=%v  unknowns=%d\n", intSum, strs, unknowns)

	fmt.Println("== Challenge: LinkedList ==")
	var l LinkedList
	for _, v := range []int{1, 2, 3, 4, 5} {
		l.Append(v)
	}
	fmt.Println("  initial:", l.ToSlice(), "len:", l.Len())

	deleted := l.Delete(3)
	fmt.Printf("  after Delete(3): %v  deleted=%t\n", l.ToSlice(), deleted)

	l.Prepend(0)
	fmt.Println("  after Prepend(0):", l.ToSlice())

	l.Append(6)
	fmt.Println("  after Append(6):", l.ToSlice(), "len:", l.Len())

	// Edge case: delete head
	l.Delete(0)
	fmt.Println("  after Delete(0):", l.ToSlice())

	// Edge case: delete non-existent
	notFound := l.Delete(99)
	fmt.Printf("  Delete(99) returned: %t\n", notFound)
}

// Day 03 walkthrough — run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson. Run it, change things,
// break things, re-run. Breaking is learning.
package main

import (
	"fmt"
	"time"
)

// ---- Structs ---------------------------------------------------------------

type Address struct {
	Street string
	City   string
	Zip    string
}

type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
	Address   Address // embedded by value — copied when User is copied
}

// ---- Embedding / Composition -----------------------------------------------

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

type Dog struct {
	Animal        // embedded — Name and Speak() are promoted
	Breed  string
}

// Dog can override the promoted method:
func (d Dog) Speak() string {
	return d.Name + " barks"  // d.Name promoted from Animal
}

// ---- Value vs Pointer receivers --------------------------------------------

type Counter struct {
	n int
}

// Value receiver — gets a copy; cannot mutate the original.
func (c Counter) Value() int {
	return c.n
}

// Pointer receiver — gets address; CAN mutate.
func (c *Counter) Increment() {
	c.n++
}

func (c *Counter) Reset() {
	c.n = 0
}

// ---- Pointer basics --------------------------------------------------------

func incrementByValue(n int) {
	n++ // modifies local copy only
}

func incrementByPointer(n *int) {
	*n++ // dereferences and modifies original
}

// Returning pointer to local — safe in Go (GC manages lifetime).
func newCounter(start int) *Counter {
	c := Counter{n: start} // escapes to heap
	return &c              // safe: Go GC tracks it
}

// ---- Control flow ----------------------------------------------------------

// Type switch — common with interfaces
func describe(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("int(%d)", v)
	case float64:
		return fmt.Sprintf("float64(%.2f)", v)
	case string:
		return fmt.Sprintf("string(%q)", v)
	case bool:
		return fmt.Sprintf("bool(%t)", v)
	default:
		return fmt.Sprintf("unknown(%T)", v)
	}
}

// If-init pattern — scope-limited variable
func findUser(id int) (User, bool) {
	users := map[int]User{
		1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
		2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
	}
	u, ok := users[id]
	return u, ok
}

func main() {
	// ---- 1. Struct initialization -------------------------------------------
	fmt.Println("== Struct initialization ==")

	// Named fields (preferred — robust to field additions)
	u := User{
		ID:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Address: Address{
			Street: "123 Main St",
			City:   "Bangalore",
			Zip:    "560001",
		},
	}
	fmt.Printf("  User: %+v\n", u) // %+v shows field names

	// Zero value — all fields zeroed
	var empty User
	fmt.Printf("  Zero User: ID=%d Name=%q Email=%q\n", empty.ID, empty.Name, empty.Email)

	// Struct is a value type — copying is explicit
	u2 := u
	u2.Name = "Modified"
	fmt.Printf("  original=%q  copy=%q (independent)\n", u.Name, u2.Name)

	// ---- 2. Embedding -------------------------------------------------------
	fmt.Println("== Embedding / composition ==")
	d := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Labrador",
	}
	fmt.Println(" ", d.Name)     // promoted from Animal
	fmt.Println(" ", d.Speak())  // Dog's own Speak() — shadowed
	fmt.Println(" ", d.Animal.Speak()) // explicit path to Animal.Speak()
	fmt.Println(" ", d.Breed)    // Dog's own field

	// ---- 3. Value vs Pointer receivers --------------------------------------
	fmt.Println("== Value vs Pointer receivers ==")
	c := Counter{}
	fmt.Printf("  initial value: %d\n", c.Value())
	c.Increment() // Go auto-converts: (&c).Increment()
	c.Increment()
	c.Increment()
	fmt.Printf("  after 3 increments: %d\n", c.Value())
	c.Reset()
	fmt.Printf("  after reset: %d\n", c.Value())

	// Value receiver does NOT mutate — even through a pointer
	cp := newCounter(10)
	before := cp.Value()
	cp.Increment()
	fmt.Printf("  newCounter(10): before=%d after=%d\n", before, cp.Value())

	// ---- 4. Pointer basics --------------------------------------------------
	fmt.Println("== Pointers ==")
	x := 42
	p := &x
	fmt.Printf("  x=%d  p=%p  *p=%d\n", x, p, *p)

	incrementByValue(x)
	fmt.Printf("  after incrementByValue: x=%d  (unchanged)\n", x)

	incrementByPointer(&x)
	fmt.Printf("  after incrementByPointer: x=%d  (changed)\n", x)

	// nil pointer guard
	var up *User
	if up == nil {
		fmt.Println("  *User is nil — guarded before dereference")
	}

	// ---- 5. if-init pattern -------------------------------------------------
	fmt.Println("== if-init pattern ==")
	if u, ok := findUser(1); ok {
		fmt.Printf("  found: %s <%s>\n", u.Name, u.Email)
	} else {
		fmt.Println("  not found")
	}
	if _, ok := findUser(99); !ok {
		fmt.Println("  user 99: not found (correct)")
	}
	// u is NOT accessible here — scoped to the if block. Clean.

	// ---- 6. Switch --------------------------------------------------------
	fmt.Println("== switch ==")

	status := "pending"
	switch status {
	case "pending":
		fmt.Println("  processing pending status")
	case "done", "cancelled":
		fmt.Println("  archiving done/cancelled")
	default:
		fmt.Printf("  unknown: %s\n", status)
	}

	// Type switch
	fmt.Println("== type switch ==")
	values := []interface{}{42, 3.14, "hello", true, []int{1, 2}}
	for _, v := range values {
		fmt.Println(" ", describe(v))
	}

	// ---- 7. for as while, range Unicode -----------------------------------
	fmt.Println("== for as while ==")
	n := 16
	doublings := 0
	for n > 1 {
		n /= 2
		doublings++
	}
	fmt.Printf("  16 halved %d times to reach 1\n", doublings)

	fmt.Println("== range over string (Unicode) ==")
	word := "héllo"
	fmt.Printf("  len(%q) = %d bytes  (not %d chars)\n", word, len(word), 5)
	for i, r := range word {
		fmt.Printf("  byte-index=%-3d rune=%c\n", i, r)
	}
	// Note: 'é' is 2 bytes → index jumps from 1 to 3

	// ---- 8. Labeled break ---------------------------------------------------
	fmt.Println("== labeled break (exits outer loop) ==")
outer:
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i == 2 && j == 2 {
				fmt.Printf("  found target at (%d,%d) — breaking outer\n", i, j)
				break outer
			}
		}
	}
	fmt.Println("  after labeled break")

	// ---- 9. Go 1.22 loop variable semantics --------------------------------
	fmt.Println("== Go 1.22 loop variable fix ==")
	funcs := make([]func(), 3)
	// In Go 1.22+, each iteration has its own `i` — prints 0, 1, 2.
	// In Go 1.21-, all closures would share the same `i` and print 3, 3, 3.
	for i := 0; i < 3; i++ {
		i := i // pre-1.22 workaround (still safe in 1.22+, just redundant)
		funcs[i] = func() { fmt.Printf("  closure: i=%d\n", i) }
	}
	for _, f := range funcs {
		f()
	}

	// ---- 10. Struct tags (preview for Day 5) --------------------------------
	fmt.Println("== Struct tags (we'll use these heavily on Day 5) ==")
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price,omitempty"`
	}
	prod := Product{ID: 1, Name: "Widget", Price: 9.99}
	fmt.Printf("  Product struct: %+v\n", prod)
	fmt.Println("  (tags are invisible here; json.Marshal uses them — see Day 5)")
}

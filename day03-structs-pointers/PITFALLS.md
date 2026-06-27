# Day 03 Pitfalls — Structs, Pointers, Control Flow

**Trap → Why it bites → Fix.**

---

### 1. Value receiver can't mutate the receiver

**Trap**
```go
func (a Account) Deposit(n int) { a.Balance += n } // mutates a copy
acc.Deposit(50) // acc.Balance unchanged
```

**Why it bites** A value receiver gets a copy of the struct; the change is thrown
away when the method returns.

**Fix** Use a pointer receiver for any method that mutates:
`func (a *Account) Deposit(n int)`.

---

### 2. Value receiver on a struct with a map/slice field DOES mutate

**Trap**
```go
type Cache struct{ data map[string]int }
func (c Cache) Set(k string, v int) { c.data[k] = v } // mutates the real map!
```

**Why it bites** The struct copy is shallow — the map header is copied but still
points to the same underlying map. So a "value receiver" surprisingly mutates
shared state, the opposite of pitfall #1.

**Fix** Be deliberate: use pointer receivers consistently so the mutation rules
are predictable, and know that map/slice fields are reference types.

---

### 3. Mixing value and pointer receivers on one type

**Trap**
```go
func (c Counter) Value() int   {}   // value
func (c *Counter) Inc()        {}   // pointer
```

**Why it bites** The method set differs between `Counter` and `*Counter`. Storing
a `Counter` (not `*Counter`) in an interface that needs `Inc` fails, and copies
behave inconsistently.

**Fix** Pick one. If any method needs a pointer receiver, make them all pointer
receivers.

---

### 4. Calling a pointer-receiver method on a non-addressable value

**Trap**
```go
Counter{}.Inc()          // compile error: cannot take address of Counter{}
m["k"].Inc()             // also fails — map values aren't addressable
```

**Why it bites** Go auto-takes the address only of *addressable* values
(variables, slice elements). Temporaries and map values aren't addressable.

**Fix** Assign to a variable first (`c := Counter{}; c.Inc()`), or store pointers
in the map (`map[string]*Counter`).

---

### 5. nil pointer dereference

**Trap**
```go
var u *User
fmt.Println(u.Name) // panic: nil pointer dereference
```

**Why it bites** The zero value of a pointer is `nil`; dereferencing it (or
accessing a field through it) panics — the most common Go panic.

**Fix** Guard: `if u == nil { return ErrNoUser }` before use.

---

### 6. Positional struct literals break silently on field changes

**Trap**
```go
u := User{1, "Alice", "a@x.com"} // add a field in the middle -> values shift
```

**Why it bites** A reordered or inserted field reassigns every positional value
with no compile error.

**Fix** Always use **named fields**: `User{ID: 1, Name: "Alice", Email: "a@x.com"}`.

---

### 7. Embedding is not inheritance

**Trap**
```go
type Dog struct{ Animal }
var a Animal = Dog{} // compile error: Dog is not an Animal
```

**Why it bites** Embedding *promotes* fields and methods, but a `Dog` is not a
subtype of `Animal`. You can't substitute one for the other.

**Fix** Use **interfaces** for polymorphism (Day 6). Embedding is composition /
forwarding, not subtyping.

---

### 8. `break` inside a `switch`/`select` inside a `for`

**Trap**
```go
for {
    switch x {
    case 1:
        break // breaks the switch, NOT the loop
    }
}
```

**Why it bites** `break` exits the innermost `switch`/`select`, so the loop keeps
running — often an infinite loop.

**Fix** Use a label: `Loop: for { switch ... { case 1: break Loop } }`.

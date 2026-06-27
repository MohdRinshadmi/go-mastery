# Day 03 Notes — Quick Reference

## Structs
```go
type User struct {
    ID    int
    Name  string
    Email string
}

u := User{ID: 1, Name: "Ada"}   // named fields (preferred)
u := User{1, "Ada", "a@x.com"}  // positional (fragile)
var u User                      // zero value (all fields zeroed)
p := &User{Name: "Ada"}         // pointer to struct (*User)
```
- Comparable with `==` only if all fields are comparable (no slice/map/func).
- `time.Time` zero value is year 0001, not "now"; use `*time.Time` for "unset".

## Embedding (composition, NOT inheritance)
```go
type Dog struct {
    Animal        // embedded — fields & methods promoted
    Breed string
}
d.Name      // promoted from Animal
d.Speak()   // promoted method
d.Animal.Name // explicit path on conflict
```
A `Dog` is NOT an `Animal` — use interfaces for polymorphism.

## Struct tags
```go
type Product struct {
    ID    int     `json:"id"`
    Price float64 `json:"price,omitempty"`
}
```
- Only exported fields are seen by encoding packages.
- `json:"name"` exact — no space after colon; typos silently ignored.

## Pointers
```go
x := 42
p := &x       // *int, address of x
*p = 99       // deref + assign -> x == 99
```
- No pointer arithmetic.
- Zero value is `nil`; dereferencing nil panics — guard with `if p == nil`.
- `&T{}` is idiomatic; `new(T)` rarely used for structs.
- Returning `&localVar` is safe (escape analysis + GC).

## Value vs pointer receivers
```go
func (c Counter) Value() int   { return c.n }  // read-only, small type
func (c *Counter) Increment()  { c.n++ }        // mutates -> pointer
```
- **Mutate?** Pointer receiver. **Large struct?** Pointer receiver.
- **Consistency rule:** if any method uses a pointer receiver, all should.
- Auto-addressing: `c.Increment()` -> `(&c).Increment()` for addressable `c`.
- Non-addressable (temporaries, map values) can't call pointer methods directly.

## Control flow
```go
if err := f(); err != nil { return err }   // init statement, scoped err

switch status {
case "done", "cancelled": archive()        // multi-value, no fallthrough
default: log.Print(status)
}

switch v := i.(type) {                      // type switch
case int:    ...
case string: ...
}

for i := 0; i < n; i++ {}                    // C-style
for n > 0 {}                                 // while-style
for {}                                       // infinite
for i, v := range s {}                       // range (v is a copy)
for i, r := range "héllo" {}                 // ranges RUNES; i is byte offset

outer:
for ... { for ... { break outer } }          // labeled break
```

## Go 1.22 loop variable fix
```go
// go.mod: go 1.22+  -> each iteration gets its OWN variable
for i, v := range items {
    funcs[i] = func() { fmt.Println(v) }    // correctly captures per-iteration v
}
// pre-1.22 needed: v := v
```

## Key terms
- **Value receiver / pointer receiver** — method gets a copy / the address.
- **Embedding** — anonymous field; promotes fields & methods (composition).
- **Method set** — the methods callable on `T` vs `*T`.
- **Escape analysis** — compiler decision: stack vs heap allocation.
- **Addressable** — a value whose address `&` can be taken.
- **Type switch** — `switch v := i.(type)`, branches on dynamic type.
- **Labeled break/continue** — `break label` to control an outer loop.

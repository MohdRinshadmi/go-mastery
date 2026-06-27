# Day 06 Interview Questions — Methods & Interfaces

Ten questions. The first seven come from the lesson; the last three are the kind
of follow-ups a senior interviewer drills you on. Answers are hidden — think
first, then expand.

---

### 1. What is the difference between a value receiver and a pointer receiver? When would you use each?

<details>
<summary>Answer</summary>

A value receiver `(r T)` operates on a **copy** of the value — mutations don't
escape the method. A pointer receiver `(r *T)` operates on the **original** and
can mutate it. Use a value receiver for small, immutable value types (`Point`,
`Money`, `time.Time`-style). Use a pointer receiver when the method mutates the
receiver, when the struct is large enough that copying is wasteful (more than a
couple of words), or to keep consistency once any method needs a pointer. In
production Go, pointer receivers are the default for structs.

</details>

---

### 2. Explain Go's implicit interface satisfaction. How does it differ from Java's `implements`?

<details>
<summary>Answer</summary>

In Go you never declare that a type implements an interface — if the type has the
required method set, the compiler treats it as satisfying the interface
(structural typing). Java is nominal: a class must explicitly say
`implements Animal` at definition time, so the author of the type must know about
the interface in advance. Go decouples this: you can define an interface *after*
the concrete type exists (even for stdlib or third-party types) and they satisfy
it with zero changes. Producers and consumers stay independent.

</details>

---

### 3. What is the method set of type `T` vs `*T`? Give an example where this matters.

<details>
<summary>Answer</summary>

The method set of `T` contains methods with **value** receivers. The method set of
`*T` contains methods with **both** value and pointer receivers. So a pointer can
call everything; a value can only call value-receiver methods. It matters for
interface satisfaction:
```go
type Sizer interface{ Size() int }
type Box struct{ n int }
func (b *Box) Size() int { return b.n } // pointer receiver

var s Sizer = &Box{5} // ok
var s Sizer = Box{5}   // compile error: Box lacks Size() in its method set
```

</details>

---

### 4. Why does Go favor small interfaces? What's wrong with a 10-method interface?

<details>
<summary>Answer</summary>

"The bigger the interface, the weaker the abstraction." Small interfaces
(`io.Reader`, `io.Writer`, `fmt.Stringer`, `error` — all one method) are trivial
to satisfy, mock, and compose. A 10-method interface forces every implementer and
every mock to provide ten methods, couples callers to behavior they don't use,
and is hard to satisfy with adapters. Small interfaces let one function accept
files, buffers, sockets, and HTTP bodies interchangeably. Compose big contracts
from small ones (`io.ReadWriteCloser`) rather than declaring fat interfaces.

</details>

---

### 5. Explain the nil interface gotcha. What does `(type, value)` mean internally?

<details>
<summary>Answer</summary>

An interface value is stored as two words: a **type** descriptor and a **value**
(data pointer). A true nil interface is `(nil, nil)`. If you box a nil concrete
pointer — e.g. return a nil `*MyError` from a function declared to return
`error` — you get `(*MyError, nil)`. The type word is set, so the interface is
**not** equal to `nil`, and `if err != nil` fires on a "successful" path. Fix:
return the untyped `nil` literal directly on success; never return a concrete
typed nil into an interface.

</details>

---

### 6. You have `func process(v interface{})`. When is a type assertion safe vs risky? What's the alternative?

<details>
<summary>Answer</summary>

The single-result assertion `x := v.(T)` **panics** if `v`'s dynamic type isn't
`T`, so it's only safe when you're certain of the type. The comma-ok form
`x, ok := v.(T)` is safe — `ok` is false instead of panicking. For branching on
several types, use a type switch: `switch x := v.(type) { case int: ...; }`.
Deeper point: frequent assertions are a design smell — if you keep recovering the
concrete type, the interface is too broad. Redesign it to expose the behavior you
need so the compiler checks it for you.

</details>

---

### 7. Why does Go not have inheritance? What problem does this solve in large codebases?

<details>
<summary>Answer</summary>

Go deliberately omits class inheritance and offers **composition** instead:
interface embedding and struct embedding (method promotion). This avoids the
fragile base class problem (a change to a parent silently breaks distant
children), the diamond problem of multiple inheritance, and the need to trace
behavior up a class hierarchy. Behaviors stay explicit and discoverable, and you
swap implementations at runtime via interfaces. "Favor composition over
inheritance" (Gang of Four, 1994) — Go just made it the only option.

</details>

---

### 8. What is a method value vs a method expression?

<details>
<summary>Answer</summary>

A **method value** binds the receiver now and gives you a function you can call
later: `f := t.Method` captures `t`, so `f(args)` runs `t.Method(args)`. A
**method expression** leaves the receiver unbound and makes it the first
parameter: `f := T.Method` (or `(*T).Method`) gives `f(t, args)`. Method values
are handy for callbacks/closures over a specific instance; method expressions are
handy when you want to apply a method across many receivers. Note: taking a
pointer-method value of an addressable value auto-takes its address.

</details>

---

### 9. Does struct embedding make the outer type satisfy an interface the embedded type satisfies?

<details>
<summary>Answer</summary>

Yes. Embedding **promotes** the embedded type's methods to the outer type, so the
outer type's method set includes them and it satisfies the same interface — no
explicit forwarding needed. Caveat: receiver kind still applies. If the interface
method has a pointer receiver on the embedded type, you may need `*Outer` (or to
embed `*Embedded`) for the promoted method to be in the value's method set. You
can also override a promoted method by defining one with the same name on the
outer type; the outer one shadows the embedded one.

</details>

---

### 10. Why define interfaces in the consumer package rather than the producer package?

<details>
<summary>Answer</summary>

Because the consumer knows what behavior it needs; the producer shouldn't have to
predict every future use. Defining the interface where it's used keeps it small
(only the methods that consumer calls), avoids forcing producers to import an
abstraction package, and breaks import cycles: package A exports concrete types,
package B declares the small interface it needs and accepts it. This is also how
you make code testable — the consumer's interface is exactly what your mock
implements. "Accept interfaces, return structs."

</details>

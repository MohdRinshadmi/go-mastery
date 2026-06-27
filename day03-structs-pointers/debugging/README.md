# Day 03 Debugging Challenge — Deposits that vanish

An `Account` type has `Deposit` and `Withdraw` methods. The program deposits 50
and withdraws 20 against a starting balance of 100, so the balance should end at
**130**. Instead it prints **100** — every mutation is lost.

This is the **value-receiver-can't-mutate** trap from today's lesson.

## Reproduce

```bash
cd bugged
go run .
```

Observed output:

```
Ada's balance: 100   <-- expected 130
```

## Hint

Look at the receiver in `func (a Account) Deposit(...)`. Is `a` the caller's
account, or a copy of it? What does `a.Balance += amount` actually modify?

<details>
<summary>Solution &amp; why</summary>

`Deposit` and `Withdraw` use **value receivers** (`a Account`). A value receiver
gets a *copy* of the struct. `a.Balance += amount` updates that copy's field, and
the moment the method returns the copy is discarded — the caller's `acc` is never
touched. So both calls are no-ops as far as `main` can see, and the balance stays
at 100.

**Fix:** use **pointer receivers** so the method operates on the original:

```go
func (a *Account) Deposit(amount int)  { a.Balance += amount }
func (a *Account) Withdraw(amount int) { a.Balance -= amount }
```

Because `acc` is an addressable variable, you don't change the call site — Go
automatically rewrites `acc.Deposit(50)` to `(&acc).Deposit(50)`. Now the balance
ends at 130.

Rules to internalize:
- **Need to mutate the receiver? Use a pointer receiver.** Value receivers are for
  read-only methods on small types.
- **Be consistent:** if any method on a type needs a pointer receiver, give them
  all pointer receivers (mixing causes subtle method-set/interface bugs).
- A value receiver on a struct that contains a *map or slice field* can still
  mutate that field's contents, because those are reference types — a common
  source of "why did this one work but not that one?" confusion.
</details>

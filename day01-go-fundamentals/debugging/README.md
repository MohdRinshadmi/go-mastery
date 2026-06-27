# Day 01 Debugging Challenge — "All withdrawals succeeded" (they didn't)

A small bank-account program processes a list of withdrawals against a balance.
The last withdrawal overdraws the account, so the program correctly prints a
`withdraw 40 failed` line — but then it cheerfully reports **`RESULT: all
withdrawals succeeded`**. The per-withdrawal logging and the final verdict
disagree. That contradiction is the bug.

This is the variable-**shadowing** trap from today's lesson, the same class of
bug `go vet -vet=shadow` warns about.

## Reproduce

```bash
cd bugged
go run .
```

Observed output:

```
withdrew 30, balance now 70
withdrew 50, balance now 20
withdraw 40 failed: insufficient funds
RESULT: all withdrawals succeeded   # <-- wrong, a withdrawal clearly failed
```

## Hint

Look at how `err` is declared **outside** the loop versus how it's declared
**inside** the loop. Count the `:=` operators. Which `err` does the final
`if err != nil` actually read?

<details>
<summary>Solution &amp; why</summary>

Inside the loop the line is:

```go
newBalance, err := withdraw(balance, amount) // := declares a NEW err
```

Because `newBalance` is brand-new, Go is happy to use `:=` — but `:=` declares
**every** name on its left that isn't already in *this* scope. The loop body is
a new scope, so this creates a fresh inner `err` that shadows the outer one. The
outer `err` declared with `var err error` is never written to; it stays `nil`.
After the loop, `if err != nil` reads that always-`nil` outer variable, so the
"all succeeded" branch always wins.

**Fix:** declare `newBalance` separately and *assign* to the existing `err` with
`=` (not `:=`) so the outer variable is the one being updated:

```go
var newBalance int
newBalance, err = withdraw(balance, amount) // = assigns to the outer err
```

Now the outer `err` holds the last failure and the final verdict is correct:

```
RESULT: some withdrawals failed
```

Takeaways:
- `:=` requires at least one new name on the left, but it will silently shadow
  any *other* names that already exist in an outer scope.
- When you mean "update the variable I already have," use `=`.
- Run `go vet` (and `gopls`/`go vet -vet=shadow`) — shadowing is exactly the
  kind of bug it flags.
</details>

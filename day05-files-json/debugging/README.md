# Day 05 Debugging Challenge — The JSON fields that vanish

A `Config` struct is decoded from a JSON string that clearly contains `port`,
`host`, and `timeout`. The decode succeeds with **no error** — but `host` comes
back as `""` and `timeout` as `0`. Re-encoding the config drops them entirely.
Only `port` survives the round trip.

This is the **"why is my JSON field empty?" #1 bug**: unexported struct fields.

## Reproduce

```bash
cd bugged
go run .
```

Observed output:

```
Port=8080 host="" timeout=0
re-encoded: {"port":8080}
```

## Hint

Look at the **capitalization** of the field *names* in the struct (not the json
tags). Which fields can a package *outside* `main` — like `encoding/json` — see?

> Note: `go vet` flags this one ("struct field has json tag but is not
> exported") — a hint the toolchain hands you for free. The program still
> *compiles and runs*, it's just wrong.

<details>
<summary>Solution &amp; why</summary>

`encoding/json` uses reflection, and reflection can only access **exported**
(capitalized) struct fields from another package. `host` and `timeout` start with
a lowercase letter, so they're unexported — invisible to `json`. The struct tag
(`json:"host"`) doesn't help: a tag on an unexported field does nothing. So on
decode those fields are never populated (they keep their zero values), and on
encode they're skipped — no error either way.

**Fix:** export the fields by capitalizing them. Keep the JSON keys lowercase via
the tags:

```go
type Config struct {
    Port    int    `json:"port"`
    Host    string `json:"host"`
    Timeout int    `json:"timeout"`
}
```

Now the round trip works:

```
Port=8080 Host="db.internal" Timeout=30
re-encoded: {"port":8080,"host":"db.internal","timeout":30}
```

Rules to internalize:
- **Only exported (capitalized) fields are (de)serialized.** This is the most
  common JSON bug for Go newcomers.
- The **field name's case** controls visibility; the **json tag** controls only
  the wire key name. You need both: an exported field *and* a tag for the
  lowercase key.
- `go vet` catches "json tag on unexported field" — run it.
</details>

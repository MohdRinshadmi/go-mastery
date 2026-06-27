# Day 05 Pitfalls — Files, io, JSON

**Trap → Why it bites → Fix.**

---

### 1. Unexported struct fields are invisible to JSON

**Trap**
```go
type Config struct {
    Port int    `json:"port"`
    host string `json:"host"` // lowercase -> never (de)serialized
}
```

**Why it bites** `encoding/json` (reflection) can only touch **exported**
(capitalized) fields. Unexported fields are silently skipped on decode *and*
encode — no error, just empty values. The #1 "why is my field empty?" bug.

**Fix** Capitalize the field; keep the wire key lowercase with the tag:
`Host string \`json:"host"\``. Run `go vet` — it flags this.

---

### 2. Forgetting `&` on Unmarshal/Decode

**Trap**
```go
json.Unmarshal(data, p)        // value, not pointer
json.NewDecoder(r).Decode(p)   // same mistake
```

**Why it bites** Decoding needs to *write into* your struct, so it needs a
pointer. Pass a value and you get an `InvalidUnmarshalError` (or nothing decoded).

**Fix** Pass the address: `json.Unmarshal(data, &p)`.

---

### 3. Forgetting `bufio.Writer.Flush()`

**Trap**
```go
w := bufio.NewWriter(f)
fmt.Fprintln(w, "line")
// no Flush -> data sits in the buffer, never hits disk
```

**Why it bites** A buffered writer holds bytes in memory until the buffer fills or
you flush. Without `Flush()`, your output silently disappears.

**Fix** `w.Flush()` before the function returns (and check its error).

---

### 4. `os.ReadFile` on unbounded input

**Trap**
```go
data, _ := os.ReadFile("huge.log") // loads the WHOLE file into RAM
```

**Why it bites** `ReadFile` reads everything into a single `[]byte`. On a
multi-GB file you OOM the process/pod.

**Fix** Stream with `bufio.Scanner` (or `bufio.Reader`) for unbounded/user-sized
input; reserve `ReadFile` for small, bounded config.

---

### 5. Not checking `scanner.Err()` after the loop

**Trap**
```go
for scanner.Scan() { use(scanner.Text()) }
// loop ends — was it EOF, or a read error?
```

**Why it bites** `Scan()` returns `false` on *both* end-of-input and on error.
Skip the check and you silently treat a truncated/failed read as a clean finish.

**Fix** After the loop: `if err := scanner.Err(); err != nil { return err }`.

---

### 6. `bufio.Scanner`'s default line-size limit

**Trap**
```go
scanner := bufio.NewScanner(f) // tokens > ~64 KB cause "token too long"
```

**Why it bites** The default max token size is ~64 KB. A single very long line
(e.g. a giant JSON line) makes `Scan()` stop early with an error.

**Fix** `scanner.Buffer(buf, maxSize)` to raise the limit, or use
`bufio.Reader.ReadString('\n')`.

---

### 7. `omitempty` can't tell "zero" from "absent"

**Trap**
```go
type T struct{ Count int `json:"count,omitempty"` }
// Count == 0 is omitted — but maybe 0 was a real, meaningful value
```

**Why it bites** `omitempty` drops zero values on encode and gives you a zero on a
missing key when decoding, so "present and zero" and "absent" look identical.

**Fix** Use a pointer field (`*int`) — `nil` means absent, `&0` means present-zero
— or `json.RawMessage`.

---

### 8. JSON numbers decode to `float64` in `interface{}`

**Trap**
```go
var m map[string]interface{}
json.Unmarshal(data, &m)
id := m["id"].(int) // panic: it's a float64, not an int
```

**Why it bites** When decoding into `interface{}`, all JSON numbers become
`float64`. A type assertion to `int` panics.

**Fix** Assert to `float64` (and convert), or decode into a typed struct so the
field is a real `int`.

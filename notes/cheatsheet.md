# Go Cheatsheet — Dense Quick Reference

> Skim target. Each block is a shape to copy, not a tutorial. For the *why*, see the matching day. For the traps, see [pitfalls.md](pitfalls.md).

---

## Variables, constants, iota

```go
var x int            // zero value: 0
var s = "hi"         // type inferred
y := 42              // short decl (functions only)
var a, b = 1, "two"  // multiple
const Pi = 3.14159   // untyped const, high precision
const ( _ = iota; KB = 1 << (10 * iota); MB; GB ) // iota: 0,1,2,3...
```

Zero values: `0`, `0.0`, `""`, `false`, `nil` (pointers, slices, maps, channels, funcs, interfaces). Every variable is always initialized — no "undefined".

```go
type Status int
const ( Pending Status = iota; Active; Closed ) // enum idiom
func (s Status) String() string { ... }         // make it printable
```

---

## Functions, multiple returns, variadics, closures

```go
func add(a, b int) int { return a + b }
func divmod(a, b int) (int, int) { return a / b, a % b }   // multiple return
func parse(s string) (n int, err error) { ... return }     // named returns
func sum(nums ...int) int { ... }                          // variadic
sum(xs...)                                                  // spread a slice

f := func(x int) int { return x * 2 }   // closure / func value
defer f.Close()                          // defer runs LIFO at return
```

`error` is always the **last** return value; `nil` means success.

---

## Slices & maps ops

```go
s := []int{1, 2, 3}
s = append(s, 4)               // ALWAYS reassign the result
s = append(s, other...)        // concat
clone := append([]int{}, s...) // idiomatic deep-ish copy of the header's data
b := make([]int, len(s)); copy(b, s)   // explicit independent copy
s = append(s[:i], s[i+1:]...)  // delete index i (order preserved, aliases!)
sub := s[1:3:3]                // three-index: cap-limited, append won't alias

m := map[string]int{"a": 1}
v, ok := m["a"]                // comma-ok: ok=false if absent
delete(m, "a")                 // safe even if absent
m = make(map[string]int, 100)  // pre-size hint
set := map[string]struct{}{}   // idiomatic set; set[k] = struct{}{}
```

Slice header = `{ptr, len, cap}`. `append` may reallocate (cap exceeded) or alias (cap free). Maps & slices are reference-ish: copying the variable copies the header, not the data.

---

## Structs, methods, interfaces

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    age  int    // unexported: not marshaled, package-private
}
u := User{ID: 1, Name: "Ada"}     // keyed literal (preferred)
p := &User{ID: 2}                  // pointer to struct

func (u User) Display() string  { return u.Name } // value receiver: gets a copy
func (u *User) SetName(n string){ u.Name = n }    // pointer receiver: mutates

type Stringer interface { String() string }       // satisfied implicitly
var _ Stringer = (*User)(nil)                      // compile-time assertion

// Embedding (composition, not inheritance)
type Admin struct { User; Level int }              // Admin.Name promoted
```

Interface satisfaction is **structural & implicit** — no `implements` keyword. Method set: value `T` has value-receiver methods; pointer `*T` has both. Store pointers in interfaces when methods have pointer receivers.

---

## Error handling idioms

```go
if err != nil { return fmt.Errorf("loading config: %w", err) } // wrap with %w

var ErrNotFound = errors.New("not found")           // sentinel
errors.Is(err, ErrNotFound)                          // walks wrap chain (not ==)

type ValidationError struct { Field, Msg string }
func (e *ValidationError) Error() string { ... }
var ve *ValidationError
if errors.As(err, &ve) { use(ve.Field) }             // extract typed error

errors.Join(err1, err2)                              // combine (Go 1.20+)
```

`%w` wraps (inspectable), `%v` flattens (hides cause at API boundary). Error strings: lowercase, no trailing punctuation. Never `err == ErrFoo`; use `errors.Is`.

---

## Goroutines, channels, select, context

```go
go work()                          // launch; runtime kills it when main returns
ch := make(chan int)               // unbuffered: send blocks until receiver ready
ch := make(chan int, 8)            // buffered: blocks only when full
ch <- v                            // send
v, ok := <-ch                      // receive; ok=false when closed & drained
close(ch)                          // SENDER closes; send-after-close panics
for v := range ch { ... }          // drains until closed
func gen() <-chan int { ... }      // directional: receive-only return
func sink(out chan<- int) { ... }  // send-only param

select {
case v := <-ch:        use(v)
case ch2 <- x:         // send case
case <-ctx.Done():     return ctx.Err()
case <-time.After(d):  // timeout (allocates a timer — see pitfalls)
default:               // non-blocking
}
```

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()                     // ALWAYS, even on the happy path
ctx, cancel := context.WithCancel(parent)
context.Background()               // root; context.TODO() placeholder
// ctx is the FIRST param; never store it in a struct.
```

```go
var wg sync.WaitGroup
for _, j := range jobs {
    wg.Add(1)
    go func() { defer wg.Done(); process(j) }() // Go 1.22+: j is per-iteration
}
wg.Wait()

g, ctx := errgroup.WithContext(ctx)             // golang.org/x/sync/errgroup
g.Go(func() error { return fetch(ctx) })        // first error cancels ctx
err := g.Wait()                                  // returns first non-nil error
```

Worker pool: N goroutines `range` a shared `jobs` channel, push to `results`. Fan-out: many readers off one channel. Fan-in: merge many channels into one. Pipeline: each stage is `<-chan in` → `chan out`, `defer close(out)`.

---

## defer / panic / recover

```go
defer f.Close()                    // LIFO; runs on every return path
defer func() {                     // recover only works inside a deferred func
    if r := recover(); r != nil {
        err = fmt.Errorf("recovered: %v", r)   // needs a NAMED return to set err
    }
}()
panic("unrecoverable")             // unwinds stack; un-recovered panic kills process
```

Panic is for "the world is broken" (impossible states, startup misconfig), not expected failures. The one prod use of recover: per-request/per-worker boundary so one bad request can't crash the server.

---

## Common stdlib

**fmt**
```go
fmt.Printf("%d %s %v %+v %#v %T %q %w", ...) // %+v: fields; %#v: Go-syntax; %T: type
s := fmt.Sprintf("%05.2f", x)                // formatted string
n, err := fmt.Sscanf(line, "%d-%d", &a, &b)
```

**strings**
```go
strings.Contains, HasPrefix, HasSuffix, Split, Join, ReplaceAll, TrimSpace,
ToLower, ToUpper, Fields, Repeat, Index, Count
var b strings.Builder; b.WriteString("x"); b.String()   // efficient concat
```

**strconv**
```go
strconv.Atoi("42"); strconv.Itoa(42)
strconv.ParseFloat(s, 64); strconv.ParseInt(s, 10, 64); strconv.ParseBool(s)
strconv.FormatInt(n, 16); strconv.Quote(s)
```

**encoding/json**
```go
b, err := json.Marshal(v)              // only EXPORTED fields, via struct tags
err = json.Unmarshal(b, &v)
json.NewEncoder(w).Encode(v)           // stream to io.Writer
json.NewDecoder(r).Decode(&v)          // stream from io.Reader
`json:"name,omitempty"`  `json:"-"`    // tags: rename / skip-if-zero / never
```

**net/http**
```go
http.HandleFunc("/x", func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()                 // cancels on client disconnect
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(resp)
})
http.ListenAndServe(":8080", mux)
resp, err := http.Get(url); defer resp.Body.Close()
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
srv := &http.Server{Addr: ":8080", Handler: mux}  // for graceful shutdown
srv.Shutdown(ctx)
```

**sync / sync/atomic**
```go
var mu sync.Mutex;   mu.Lock(); defer mu.Unlock()
var rw sync.RWMutex; rw.RLock()/.RUnlock(); rw.Lock()/.Unlock()
var once sync.Once;  once.Do(initFn)
var wg sync.WaitGroup
var n atomic.Int64;  n.Add(1); n.Load(); n.Store(0); n.CompareAndSwap(0, 1)
var p sync.Pool      // reuse allocations in hot paths
```

**time**
```go
time.Now(); time.Since(t); time.Sleep(d); d := 2 * time.Second
t.Format("2006-01-02 15:04:05"); time.Parse(layout, s)   // ref date: Mon Jan 2 ...
tk := time.NewTicker(d); defer tk.Stop()
tm := time.NewTimer(d); defer tm.Stop()                  // prefer over time.After in loops
ctx, cancel := context.WithTimeout(ctx, d)
```

**sort / slices / maps** (slices & maps packages: Go 1.21+)
```go
sort.Ints(xs); sort.Strings(ss); sort.Slice(xs, func(i, j int) bool { ... })
slices.Sort(xs); slices.Contains(xs, v); slices.Index(xs, v)
slices.SortFunc(xs, func(a, b T) int { return cmp.Compare(a.K, b.K) })
slices.Max, Min, Reverse, Equal, Clone, BinarySearch
maps.Keys(m); maps.Values(m); maps.Clone(m)   // Keys/Values return iterators (1.23+)
```

---

## go CLI commands

```bash
go run main.go              # compile + run, no binary kept
go build ./...              # compile all packages; binary in cwd
go install ./cmd/app        # build + put binary in $GOBIN
go test ./...               # run all tests
go test -run TestX -v       # one test, verbose
go test -race ./...         # data-race detector (CI default for concurrent pkgs)
go test -cover ./...        # coverage; -coverprofile=c.out
go test -bench=. -benchmem  # benchmarks + alloc stats
go vet ./...                # static checks (lock copy, printf args, etc.)
go fmt ./...                # gofmt (no debates)
go mod init example.com/m   # create go.mod
go mod tidy                 # add missing / drop unused deps; sync go.sum
go mod download             # fetch deps to module cache
go mod vendor               # vendor deps into ./vendor
go get pkg@v1.2.3           # add/upgrade a dependency
go work init / go work use  # multi-module workspaces

# profiling
go test -cpuprofile cpu.out -memprofile mem.out -bench=.
go tool pprof cpu.out       # interactive: top, list, web
go tool pprof http://localhost:6060/debug/pprof/profile   # live (net/http/pprof)
GODEBUG=gctrace=1 ./app     # GC trace
GOMAXPROCS=4 ./app          # cap OS threads running goroutines
```

Build tags: `//go:build linux && amd64` (first line, blank line after). `//go:build ignore` excludes a file from normal builds.

// Day 02 walkthrough — run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson. Run it, then change
// things and re-run to see what breaks. Breaking it is how you learn.
package main

import (
	"fmt"
	"sort"
)

// ---- Visibility demo (package-level) ------------------------------------
// Exported — visible to any package importing this one.
const MaxPageSize = 100

// Unexported — only visible inside this package.
const defaultTimeout = 30

// ---- Helper: print slice internals ----------------------------------------
// Uses the "reflect" trick — we'll just show len/cap manually.
func printSliceInfo(label string, s []int) {
	fmt.Printf("  %-20s len=%-3d cap=%-3d data=%v\n", label, len(s), cap(s), s)
}

func main() {
	// ---- 1. Arrays — value semantics ----------------------------------------
	fmt.Println("== Arrays: value type ==")
	a := [3]int{1, 2, 3}
	b := a        // full copy — b is a completely separate [3]int
	b[0] = 99
	fmt.Printf("  a=%v  b=%v  (modifying b did NOT change a)\n", a, b)
	// [3]int and [4]int are different types — this would not compile:
	// var c [4]int = a // COMPILE ERROR

	// ---- 2. Slices — reference into an array --------------------------------
	fmt.Println("== Slice internals: len, cap, pointer ==")
	s := []int{10, 20, 30, 40, 50}
	printSliceInfo("original s", s)

	s2 := s[1:3] // window: elements at index 1,2 — shares underlying array
	printSliceInfo("s2 = s[1:3]", s2)
	// cap of s2 = 5 - 1 = 4 (from index 1 to end of array)

	s3 := s[:3]
	printSliceInfo("s3 = s[:3]", s3)
	// cap = 5 (from index 0 to end)

	// ---- 3. THE ALIASING GOTCHA — read this carefully ----------------------
	fmt.Println("== THE aliasing gotcha ==")
	orig := []int{1, 2, 3, 4, 5}
	view := orig[1:3] // [2 3] — same backing array
	printSliceInfo("orig  before", orig)
	printSliceInfo("view  before", view)

	view[0] = 99 // modifies orig[1] — they share memory!
	fmt.Println("  After view[0] = 99:")
	printSliceInfo("orig  after ", orig) // orig[1] is now 99!
	printSliceInfo("view  after ", view)

	// The append trap: view has cap > len, append writes into orig's array
	fmt.Println("== The append-into-shared-memory trap ==")
	base := []int{1, 2, 3, 4, 5}
	sub := base[:3] // len=3, cap=5 — still pointing into base
	printSliceInfo("base before append", base)
	printSliceInfo("sub  before append", sub)

	sub = append(sub, 99) // cap > len → no new alloc → writes into base[3]!
	fmt.Println("  After sub = append(sub, 99):")
	printSliceInfo("base after append ", base) // base[3] is now 99!
	printSliceInfo("sub  after append ", sub)

	// ---- 4. Safe copying — three approaches --------------------------------
	fmt.Println("== Safe copy — three approaches ==")
	data := []int{10, 20, 30}

	// Option 1: make + copy
	copyA := make([]int, len(data))
	copy(copyA, data)
	copyA[0] = 999
	fmt.Printf("  make+copy: data=%v  copyA=%v  (independent)\n", data, copyA)

	// Option 2: append clone idiom
	copyB := append([]int{}, data...)
	copyB[0] = 888
	fmt.Printf("  append clone: data=%v  copyB=%v  (independent)\n", data, copyB)

	// Option 3: three-index slice limits cap → forces alloc on next append
	safe := data[0:2:2] // cap = 2-0 = 2, same as len
	safe = append(safe, 777)
	fmt.Printf("  three-index: data=%v  safe=%v  (data untouched)\n", data, safe)

	// ---- 5. append growth behavior ----------------------------------------
	fmt.Println("== append capacity growth ==")
	growing := make([]int, 0)
	prevCap := 0
	for i := 0; i < 10; i++ {
		growing = append(growing, i)
		if cap(growing) != prevCap {
			fmt.Printf("  len=%-3d  cap grows to %d\n", len(growing), cap(growing))
			prevCap = cap(growing)
		}
	}

	// Pre-allocation when you know size — no reallocations
	fmt.Println("== Pre-allocated slice: zero re-allocs ==")
	preallocated := make([]int, 0, 10)
	prevCap2 := cap(preallocated)
	for i := 0; i < 10; i++ {
		preallocated = append(preallocated, i)
		if cap(preallocated) != prevCap2 {
			fmt.Printf("  reallocation at len=%d!\n", len(preallocated))
			prevCap2 = cap(preallocated)
		}
	}
	fmt.Printf("  done, len=%d cap=%d — zero reallocations\n", len(preallocated), cap(preallocated))

	// ---- 6. copy with overlap (delete element) ----------------------------
	fmt.Println("== delete element at index 1 using copy ==")
	del := []int{10, 20, 30, 40, 50}
	idx := 1
	copy(del[idx:], del[idx+1:]) // shift left
	del = del[:len(del)-1]        // shrink
	fmt.Println(" ", del)         // [10 30 40 50]

	// ---- 7. Maps: basics ---------------------------------------------------
	fmt.Println("== Maps: basics ==")
	ages := map[string]int{
		"Alice": 30,
		"Bob":   25,
		"Carol": 28,
	}

	// comma-ok: ALWAYS use this for reads where absence matters
	if age, ok := ages["Bob"]; ok {
		fmt.Printf("  Bob is %d\n", age)
	}
	if _, ok := ages["Dave"]; !ok {
		fmt.Println("  Dave not found (comma-ok correctly detected absence)")
	}

	// Zero-value read trap
	unknownAge := ages["Dave"] // returns 0, not "not found"
	fmt.Printf("  ages[\"Dave\"] without comma-ok = %d  ← dangerous if 0 is a valid age\n", unknownAge)

	// ---- 8. Maps: sorted iteration ----------------------------------------
	fmt.Println("== Sorted map iteration ==")
	keys := make([]string, 0, len(ages))
	for k := range ages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, ages[k])
	}

	// ---- 9. Map as set (idiomatic) ----------------------------------------
	fmt.Println("== Map as set (struct{} value = zero memory) ==")
	seen := make(map[string]struct{})
	words := []string{"go", "is", "great", "go", "is"}
	for _, w := range words {
		seen[w] = struct{}{}
	}
	fmt.Printf("  unique words: %d\n", len(seen)) // 3

	_, exists := seen["go"]
	fmt.Printf("  'go' in set: %t\n", exists)

	// ---- 10. Map is a reference type ---------------------------------------
	fmt.Println("== Map reference semantics ==")
	m1 := map[string]int{"x": 1}
	m2 := m1      // m2 and m1 point to SAME underlying map
	m2["y"] = 2
	fmt.Printf("  m1=%v  (mutated through m2 — they share memory)\n", m1)

	// ---- Visibility reminder -----------------------------------------------
	fmt.Println("== Visibility (exported vs unexported) ==")
	fmt.Printf("  MaxPageSize (exported)   = %d\n", MaxPageSize)
	fmt.Printf("  defaultTimeout (unexported) = %d (accessible within same package)\n", defaultTimeout)
}

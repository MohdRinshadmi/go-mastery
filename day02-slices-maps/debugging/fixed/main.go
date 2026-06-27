package main

import "fmt"

// firstThree returns the first three readings plus a truncation marker, WITHOUT
// touching the caller's slice.
func firstThree(readings []int) []int {
	// FIX option A — three-index slice limits cap to len, so the next append
	// is forced to allocate a fresh backing array instead of overwriting
	// readings[3]:
	//
	//   preview := readings[:3:3]
	//
	// FIX option B (used here) — copy into an independent slice. Clearest intent:
	preview := make([]int, 3)
	copy(preview, readings[:3]) // preview now owns its own memory
	preview = append(preview, 999)
	return preview
}

func main() {
	readings := []int{10, 20, 30, 40, 50}

	preview := firstThree(readings)

	fmt.Println("preview:", preview)   // [10 20 30 999]
	fmt.Println("readings:", readings) // [10 20 30 40 50] — untouched

	total := 0
	for _, r := range readings {
		total += r
	}
	fmt.Println("sum of readings:", total) // 150
}

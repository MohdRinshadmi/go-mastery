package main

import "fmt"

// firstThree returns the first three readings from a sensor log so a caller can
// display a "preview", while keeping the full log for later processing.
//
// The team also tags the preview with a sentinel value 999 appended on the end
// to mark "preview truncated here".
func firstThree(readings []int) []int {
	preview := readings[:3] // share the backing array of readings
	// Append a truncation marker to the preview.
	preview = append(preview, 999) // BUG: cap(preview) > len(preview),
	//                                so this writes 999 into readings[3]!
	return preview
}

func main() {
	readings := []int{10, 20, 30, 40, 50}

	preview := firstThree(readings)

	fmt.Println("preview:", preview)   // [10 20 30 999]
	fmt.Println("readings:", readings) // EXPECTED [10 20 30 40 50]
	//                                   ACTUAL   [10 20 30 999 50] — corrupted!

	// Downstream code trusts that readings still holds the original sensor data.
	total := 0
	for _, r := range readings {
		total += r
	}
	fmt.Println("sum of readings:", total) // expected 150, but prints 1109
}

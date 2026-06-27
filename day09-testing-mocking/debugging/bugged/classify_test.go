package bugged

import "testing"

// TestClassify LOOKS like a proper table-driven test, but it is a
// false-green test. Read the assertion carefully:
//
//	if got == "" { ... }
//
// It only checks that Classify returned *some* non-empty string. It
// never compares `got` against the per-case `tc.want`. Since Classify
// always returns one of "fail"/"pass"/"distinction" (never ""), EVERY
// case passes — including the boundary case (score == 70) that the
// buggy Classify mislabels as "pass" instead of "distinction".
//
// The `want` column is collected but never asserted. Result:
// `go test ./...` is GREEN, yet Classify(70) ships wrong.
func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"low fail", 10, "fail"},
		{"just below pass", 39, "fail"},
		{"bottom of pass", 40, "pass"},
		{"mid pass", 55, "pass"},
		{"boundary distinction", 70, "distinction"}, // the case that should fail
		{"clear distinction", 95, "distinction"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.input)

			// BUG: this asserts almost nothing. It should be
			//     if got != tc.want { ... }
			// Instead it just checks the result is non-empty, so the
			// table's `want` column is never actually used.
			if got == "" {
				t.Errorf("Classify(%d) returned empty string", tc.input)
			}
		})
	}
}

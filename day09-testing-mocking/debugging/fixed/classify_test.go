package fixed

import "testing"

// TestClassify is a real table-driven test: it asserts `got` against
// the per-case `tc.want`. With the off-by-one corrected in classify.go,
// every case — including the boundary score == 70 — passes for the
// right reason.
//
// (If you paste this assertion into the bugged package WITHOUT fixing
// the off-by-one, the "boundary distinction" case goes RED — which is
// exactly the failure the false-green test was hiding.)
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
		{"top of pass", 69, "pass"},
		{"boundary distinction", 70, "distinction"}, // the case the bug got wrong
		{"clear distinction", 95, "distinction"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.input)
			if got != tc.want { // assert against the table, not a constant
				t.Errorf("Classify(%d) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

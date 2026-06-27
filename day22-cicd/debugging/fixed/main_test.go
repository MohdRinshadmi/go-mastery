package main

import "testing"

// TestSummarize now passes deterministically because summarize() sorts its
// keys. Run `go test -count=100 .` — it is green every time.
func TestSummarize(t *testing.T) {
	flags := map[string]string{
		"beta":   "on",
		"cache":  "off",
		"region": "eu",
	}
	got := summarize(flags)
	want := "beta=on,cache=off,region=eu"
	if got != want {
		t.Fatalf("summarize() = %q, want %q", got, want)
	}
}

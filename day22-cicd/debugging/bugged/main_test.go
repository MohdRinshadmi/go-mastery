package main

import "testing"

// TestSummarize is the FLAKY test. It asserts one specific ordering, but
// summarize() ranges a map (random order), so this fails whenever the
// runtime happens to pick a different order — which it eventually will in
// CI. Run `go test -count=20 .` to watch it flap.
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

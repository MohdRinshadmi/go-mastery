package bench

import (
	"strings"
	"testing"
)

const N = 1000

var sink []int // prevents dead-code elimination

func BenchmarkBuildNaive(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = BuildNaive(N)
	}
}

func BenchmarkBuildPrealloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = BuildPrealloc(N)
	}
}

var strSink string

// The classic += in a loop vs strings.Builder.
func BenchmarkConcatPlus(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 100; j++ {
			s += "x" // each += allocates a new string
		}
		strSink = s
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		for j := 0; j < 100; j++ {
			sb.WriteByte('x')
		}
		strSink = sb.String()
	}
}

// A normal correctness test still lives alongside benchmarks.
func TestBuildEquivalent(t *testing.T) {
	a, b := BuildNaive(50), BuildPrealloc(50)
	if len(a) != len(b) {
		t.Fatalf("len mismatch %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}

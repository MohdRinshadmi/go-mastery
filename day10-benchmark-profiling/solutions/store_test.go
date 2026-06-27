package store

import (
	"strconv"
	"testing"
)

// ONE test suite, run against BOTH implementations — proves substitutability.
func testStore(t *testing.T, s Store) {
	t.Helper()
	if _, ok := s.Get("missing"); ok {
		t.Error("empty store returned ok for missing key")
	}
	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("a", "10") // overwrite
	if v, ok := s.Get("a"); !ok || v != "10" {
		t.Errorf("Get(a) = %q,%v; want 10,true", v, ok)
	}
	if v, ok := s.Get("b"); !ok || v != "2" {
		t.Errorf("Get(b) = %q,%v; want 2,true", v, ok)
	}
}

func TestMapStore(t *testing.T)   { testStore(t, NewMapStore()) }
func TestSliceStore(t *testing.T) { testStore(t, NewSliceStore()) }

// Comparative benchmarks: Get on each as N grows. Map stays flat; slice
// scan grows linearly. Run with -benchmem and read ns/op across sizes.
func benchGet(b *testing.B, s Store, n int) {
	for i := 0; i < n; i++ {
		s.Set(strconv.Itoa(i), "v")
	}
	target := strconv.Itoa(n - 1) // worst case for the slice (last element)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(target)
	}
}

func BenchmarkMapGet_100(b *testing.B)    { benchGet(b, NewMapStore(), 100) }
func BenchmarkMapGet_10000(b *testing.B)  { benchGet(b, NewMapStore(), 10000) }
func BenchmarkSliceGet_100(b *testing.B)  { benchGet(b, NewSliceStore(), 100) }
func BenchmarkSliceGet_10000(b *testing.B) { benchGet(b, NewSliceStore(), 10000) }

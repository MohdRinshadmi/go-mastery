package store

import "testing"

// TODO EXERCISE: write ONE suite and run it against BOTH stores.
//   func testStore(t *testing.T, s Store) { ... set/get/overwrite/missing ... }
//   func TestMapStore(t *testing.T)   { testStore(t, NewMapStore()) }
//   func TestSliceStore(t *testing.T) { testStore(t, NewSliceStore()) }

// TODO BENCHMARK: benchGet(b, store, n) that fills n keys then Gets the
// worst-case key b.N times. Add BenchmarkMapGet_100/_10000 and
// BenchmarkSliceGet_100/_10000. Compare ns/op across sizes — that's the lesson.

func TestPlaceholder(t *testing.T) {
	// Remove once you add real tests above.
	_ = NewMapStore()
	_ = NewSliceStore()
}

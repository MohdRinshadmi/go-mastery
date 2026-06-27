package finalexam

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGroupBy(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want map[bool][]int
	}{
		{"even/odd", []int{1, 2, 3, 4}, map[bool][]int{true: {2, 4}, false: {1, 3}}},
		{"empty", nil, map[bool][]int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupBy(tc.in, func(n int) bool { return n%2 == 0 })
			if len(got) != len(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if len(got[k]) != len(v) {
					t.Errorf("group %v: got %v want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestLRU(t *testing.T) {
	c := NewLRU(2)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3") // evicts "a" (oldest)
	if _, ok := c.Get("a"); ok {
		t.Error("a should have been evicted")
	}
	if v, ok := c.Get("c"); !ok || v != "3" {
		t.Errorf("c = %q,%v", v, ok)
	}
}

func TestLRURace(t *testing.T) {
	c := NewLRU(50)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				c.Set("k", "v")
				c.Get("k")
			}
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestCheckURLs(t *testing.T) {
	check := func(ctx context.Context, url string) error {
		d := 5 * time.Millisecond
		if url == "slow" {
			d = 200 * time.Millisecond
		}
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	got := CheckURLs(context.Background(), []string{"fast", "slow"}, 2, 50*time.Millisecond, check)
	if got["fast"] != nil {
		t.Errorf("fast should be ok, got %v", got["fast"])
	}
	if !errors.Is(got["slow"], context.DeadlineExceeded) {
		t.Errorf("slow should time out, got %v", got["slow"])
	}
}

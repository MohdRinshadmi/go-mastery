package calc

import (
	"errors"
	"testing"
)

// Table-driven test — the Go idiom.
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"simple", 10, 2, 5, false},
		{"negative", -6, 3, -2, false},
		{"by zero", 1, 0, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Divide(tc.a, tc.b)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if !errors.Is(err, ErrDivByZero) {
					t.Errorf("expected ErrDivByZero, got %v", err)
				}
				return
			}
			if got != tc.want {
				t.Errorf("Divide(%v,%v) = %v; want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// Hand-written fake implementing Charger — no network, full control.
type fakeCharger struct {
	id      string
	wantErr bool
}

func (f fakeCharger) Charge(amount int) (string, error) {
	if f.wantErr {
		return "", errors.New("gateway down")
	}
	return f.id, nil
}

func TestCheckout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		txn, err := Checkout(fakeCharger{id: "txn_123"}, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if txn != "txn_123" {
			t.Errorf("got %q", txn)
		}
	})
	t.Run("invalid amount", func(t *testing.T) {
		if _, err := Checkout(fakeCharger{}, 0); err == nil {
			t.Error("expected error for amount 0")
		}
	})
	t.Run("gateway failure", func(t *testing.T) {
		if _, err := Checkout(fakeCharger{wantErr: true}, 50); err == nil {
			t.Error("expected gateway error")
		}
	})
}

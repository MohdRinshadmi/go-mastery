package wallet

import (
	"errors"
	"testing"
)

func TestWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		balance int
		amount  int
		want    int
		wantErr bool
		errIs   error // optional specific sentinel
	}{
		{"normal", 100, 30, 70, false, nil},
		{"exact balance", 50, 50, 0, false, nil},
		{"over balance", 20, 50, 20, true, ErrInsufficient},
		{"zero amount", 100, 0, 100, true, nil},
		{"negative amount", 100, -5, 100, true, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Withdraw(tc.balance, tc.amount)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.errIs != nil && !errors.Is(err, tc.errIs) {
				t.Errorf("expected errors.Is %v, got %v", tc.errIs, err)
			}
			if got != tc.want {
				t.Errorf("balance = %d, want %d", got, tc.want)
			}
		})
	}
}

type fakeCharger struct {
	id      string
	wantErr bool
	called  bool
}

func (f *fakeCharger) Charge(amount int) (string, error) {
	f.called = true
	if f.wantErr {
		return "", errors.New("gateway down")
	}
	return f.id, nil
}

func TestCheckout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fc := &fakeCharger{id: "txn_9"}
		txn, err := Checkout(fc, 100)
		if err != nil || txn != "txn_9" {
			t.Fatalf("got (%q, %v)", txn, err)
		}
	})
	t.Run("invalid amount does not call gateway", func(t *testing.T) {
		fc := &fakeCharger{}
		if _, err := Checkout(fc, 0); err == nil {
			t.Error("expected error")
		}
		if fc.called {
			t.Error("gateway should not be called for invalid amount")
		}
	})
	t.Run("gateway failure propagates", func(t *testing.T) {
		fc := &fakeCharger{wantErr: true}
		if _, err := Checkout(fc, 50); err == nil {
			t.Error("expected gateway error")
		}
	})
}

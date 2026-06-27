package wallet

import "testing"

// =====================================================================
// EXERCISE 1 — table-driven test for Withdraw
// Cover: normal withdrawal, exact-balance, over-balance (ErrInsufficient),
// zero/negative amount. Use t.Run subtests and check wantErr.
// =====================================================================
func TestWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		balance int
		amount  int
		want    int
		wantErr bool
	}{
		// TODO: add your cases here
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call Withdraw, assert error + value
			_ = tc
		})
	}
}

// =====================================================================
// EXERCISE 2 — hand-written fake Charger + Checkout tests
// Define a fakeCharger type implementing Charger. Test:
//   - success path returns the txn id
//   - amount <= 0 returns an error (gateway NOT called)
//   - gateway failure propagates the error
// =====================================================================

// TODO: type fakeCharger struct { ... }
// TODO: func (f fakeCharger) Charge(amount int) (string, error) { ... }

func TestCheckout(t *testing.T) {
	// TODO
}

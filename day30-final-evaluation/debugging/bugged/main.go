// Day 30 debugging — goroutine leak in the order -> payment wiring.
//
// The OrderService accepts an order and, for each one, fires a goroutine that
// calls the (slow) PaymentService and reports the result back on an UNBUFFERED
// channel. The caller waits for the result but gives up after a timeout (a normal
// resilience pattern, Day 25). When it gives up, it stops reading the result
// channel.
//
// BUG: the payment goroutine eventually finishes and tries to send its result on
// the unbuffered channel — but nobody is receiving anymore, so the send blocks
// FOREVER. The goroutine leaks. Under load (every slow payment leaks one), the
// process slowly accumulates goroutines until it falls over.
//
// We place many orders whose payments are slower than the timeout and then count
// leaked goroutines.
//
// Stdlib only. Run with the race detector:  go run -race .
package main

import (
	"fmt"
	"runtime"
	"time"
)

type PaymentResult struct {
	OrderID string
	OK      bool
}

// PaymentService is slow (a downstream you don't control).
func charge(orderID string) PaymentResult {
	time.Sleep(60 * time.Millisecond)
	return PaymentResult{OrderID: orderID, OK: true}
}

// placeOrder fires the payment in a goroutine and waits up to `timeout` for it.
func placeOrder(orderID string, timeout time.Duration) (PaymentResult, bool) {
	resultCh := make(chan PaymentResult) // BUG: unbuffered

	go func() {
		// When this finishes after the timeout, the caller has stopped reading,
		// so this send blocks forever -> leaked goroutine.
		resultCh <- charge(orderID)
	}()

	select {
	case res := <-resultCh:
		return res, true
	case <-time.After(timeout):
		return PaymentResult{}, false // give up; STOP reading resultCh
	}
}

func main() {
	before := runtime.NumGoroutine()

	const orders = 200
	// Timeout shorter than the payment latency, so every payment "times out"
	// and every payment goroutine leaks.
	for i := 0; i < orders; i++ {
		_, ok := placeOrder(fmt.Sprintf("order-%d", i), 10*time.Millisecond)
		_ = ok
	}

	// Give the slow payments time to finish and block on the dead channel.
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()

	fmt.Printf("goroutines before: %d, after: %d\n", before, after)
	leaked := after - before
	if leaked > 10 {
		fmt.Printf("BUG: ~%d goroutines leaked (blocked sending on an abandoned channel)\n", leaked)
	} else {
		fmt.Println("OK: no significant goroutine leak")
	}
}

// Day 30 debugging — FIXED: the payment goroutine can always make progress.
//
// The leak happened because the result channel was unbuffered, so a late sender
// blocked forever once the caller stopped receiving. The minimal, robust fix is
// a BUFFERED channel of capacity 1: the goroutine's single send always succeeds
// (the buffer slot is always free), so it returns and is reclaimed even when the
// caller has already timed out and walked away.
//
// (A complementary production fix is to give the goroutine a context so it can
// abort the downstream work on timeout; here we just ensure it can never block
// on the send, which is what fixes the leak.)
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

func charge(orderID string) PaymentResult {
	time.Sleep(60 * time.Millisecond)
	return PaymentResult{OrderID: orderID, OK: true}
}

func placeOrder(orderID string, timeout time.Duration) (PaymentResult, bool) {
	resultCh := make(chan PaymentResult, 1) // FIX: buffered (cap 1)

	go func() {
		// The buffered slot is always available, so this send never blocks even
		// if the caller already timed out. The goroutine finishes and is reclaimed.
		resultCh <- charge(orderID)
	}()

	select {
	case res := <-resultCh:
		return res, true
	case <-time.After(timeout):
		return PaymentResult{}, false // give up; the goroutine can still send & exit
	}
}

func main() {
	before := runtime.NumGoroutine()

	const orders = 200
	for i := 0; i < orders; i++ {
		_, ok := placeOrder(fmt.Sprintf("order-%d", i), 10*time.Millisecond)
		_ = ok
	}

	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()

	fmt.Printf("goroutines before: %d, after: %d\n", before, after)
	leaked := after - before
	if leaked <= 10 {
		fmt.Println("OK: no goroutine leak — late senders complete into the buffered channel and exit")
	} else {
		fmt.Printf("unexpected: ~%d goroutines still around\n", leaked)
	}
}

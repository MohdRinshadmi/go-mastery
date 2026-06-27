// Day 27 debugging — FIXED: idempotent consumer dedupes on EventID.
//
// At-least-once delivery is a fact of life — you can't make the broker stop
// redelivering. The fix lives in the CONSUMER: it records every EventID it has
// processed and skips duplicates, so the effect happens exactly once per logical
// event even when the message arrives 2+ times.
//
// Stdlib only. Run with the race detector:  go run -race .
package main

import (
	"fmt"
	"sync"
)

type OrderPaid struct {
	EventID string
	OrderID string
	Amount  int // cents
}

type broker struct {
	ch        chan OrderPaid
	redeliver map[string]bool
}

func newBroker() *broker {
	return &broker{ch: make(chan OrderPaid, 16), redeliver: map[string]bool{}}
}

func (b *broker) publish(e OrderPaid) {
	b.ch <- e
	if b.redeliver[e.EventID] {
		b.ch <- e
	}
}

type ledger struct {
	mu      sync.Mutex
	charged map[string]int
}

func (l *ledger) charge(orderID string, cents int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.charged[orderID] += cents
}

// processedStore is the idempotency guard: a set of EventIDs we've handled.
// In production this is a Redis SET / DB unique key with a TTL, not a map.
type processedStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

// markIfNew atomically records the event and reports whether it was new.
// Doing "check + record" under one lock avoids a race where two deliveries
// both pass the check before either records.
func (p *processedStore) markIfNew(eventID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.seen[eventID] {
		return false
	}
	p.seen[eventID] = true
	return true
}

func main() {
	b := newBroker()
	led := &ledger{charged: map[string]int{}}
	processed := &processedStore{seen: map[string]bool{}}

	b.redeliver["evt-1"] = true

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range b.ch {
			// FIX: idempotency guard. Skip if we've already processed this event.
			if !processed.markIfNew(e.EventID) {
				fmt.Printf("skipped duplicate event %s\n", e.EventID)
				continue
			}
			led.charge(e.OrderID, e.Amount)
			fmt.Printf("charged order %s $%.2f (event %s)\n",
				e.OrderID, float64(e.Amount)/100, e.EventID)
		}
	}()

	b.publish(OrderPaid{EventID: "evt-1", OrderID: "order-1", Amount: 4999})
	close(b.ch)
	wg.Wait()

	total := led.charged["order-1"]
	fmt.Printf("order-1 total charged: $%.2f\n", float64(total)/100)
	if total == 4999 {
		fmt.Println("OK: charged exactly once despite redelivery")
	}
}

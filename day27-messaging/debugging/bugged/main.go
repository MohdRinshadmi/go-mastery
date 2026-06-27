// Day 27 debugging — at-least-once delivery without idempotency = double charge.
//
// We simulate a message broker (Kafka/RabbitMQ stand-in) with a channel. Real
// brokers guarantee AT-LEAST-ONCE delivery: if a consumer crashes or a network
// blip happens between processing and acking, the broker REDELIVERS the message.
// Here the broker deliberately redelivers one event (as if the first ack was
// lost) so you can see what an at-least-once world does to a non-idempotent
// consumer.
//
// The payment consumer is NOT idempotent: it charges the customer every time it
// sees an OrderPaid event. Under redelivery, the customer is charged twice.
//
// Stdlib only. Run with the race detector:  go run -race .
package main

import (
	"fmt"
	"sync"
)

type OrderPaid struct {
	EventID string // unique per logical event (the dedup key — unused by the bug)
	OrderID string
	Amount  int // cents
}

// broker is a tiny at-least-once message bus: it delivers each published event,
// and redelivers any event whose EventID is in `redeliver` (simulating a lost ack).
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
		b.ch <- e // at-least-once: the same event arrives again
	}
}

// ledger is the customer's account; charges accumulate here.
type ledger struct {
	mu      sync.Mutex
	charged map[string]int // orderID -> total cents charged
}

func (l *ledger) charge(orderID string, cents int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.charged[orderID] += cents // BUG: applied once per delivery, not once per order
}

func main() {
	b := newBroker()
	led := &ledger{charged: map[string]int{}}

	// Simulate a lost ack for this event: the broker will deliver it twice.
	b.redeliver["evt-1"] = true

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range b.ch {
			// BUG: no idempotency guard. Every delivery charges the card.
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
	if total != 4999 {
		fmt.Printf("BUG: order-1 was charged twice (expected $49.99, got $%.2f)\n",
			float64(total)/100)
	}
}

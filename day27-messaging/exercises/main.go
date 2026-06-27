// Day 27 — YOUR exercises. Run: go run .
package main

import (
	"fmt"
	"sync"
)

type Event struct {
	ID      string
	Type    string
	Payload map[string]any
}
type Handler func(Event) error

// In-memory bus that delivers each event TWICE (at-least-once simulation).
type Bus struct {
	mu   sync.Mutex
	subs map[string][]Handler
}

func NewBus() *Bus { return &Bus{subs: map[string][]Handler{}} }
func (b *Bus) Subscribe(topic string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[topic] = append(b.subs[topic], h)
}
func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	hs := append([]Handler(nil), b.subs[e.Type]...)
	b.mu.Unlock()
	for _, h := range hs {
		_ = h(e)
		_ = h(e) // duplicate
	}
}

// =====================================================================
// TASK 1 — idempotent inventory consumer
// Subscribe to "OrderPlaced". Decrement a stock counter by payload["qty"],
// but ONLY ONCE per event ID (dedup on e.ID). Print when a duplicate is
// ignored. Prove final stock is correct despite double-delivery.
//
// TASK 2 — second independent consumer ("email") that prints a confirmation
// once per event ID.
//
// CHALLENGE — add a "dead-letter" path: a consumer that returns an error
// for a specific bad event id; collect failed events into a deadLetters
// slice instead of losing them.
// =====================================================================

func main() {
	bus := NewBus()
	stock := 100
	_ = stock
	_ = bus // TODO: remove once you use bus.Subscribe / bus.Publish

	// TODO: implement the consumers and publish a couple of OrderPlaced events
	fmt.Println("TODO: wire idempotent consumers; see lesson + ../solutions")
}

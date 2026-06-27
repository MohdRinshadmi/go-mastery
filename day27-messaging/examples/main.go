// Day 27 walkthrough — event-driven architecture with an in-memory bus
// (a stand-in for Kafka/RabbitMQ so it runs offline). Run: go run .
//
// Real Kafka/RabbitMQ code: ../solutions/broker_reference.go (build-ignored).
package main

import (
	"fmt"
	"sync"
)

// ---- Event + in-memory bus (pub/sub) ------------------------------------
type Event struct {
	ID      string // unique event id -> idempotency key
	Type    string
	Payload map[string]any
}

type Handler func(Event) error

type Bus struct {
	mu          sync.Mutex
	subscribers map[string][]Handler // topic -> handlers (consumer groups)
}

func NewBus() *Bus { return &Bus{subscribers: map[string][]Handler{}} }

func (b *Bus) Subscribe(topic string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], h)
}

// Publish fans out to all subscribers. Deliberately delivers TWICE to
// simulate at-least-once delivery — your handlers must be idempotent.
func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	handlers := append([]Handler(nil), b.subscribers[e.Type]...)
	b.mu.Unlock()
	for _, h := range handlers {
		_ = h(e) // first delivery
		_ = h(e) // duplicate delivery (at-least-once reality)
	}
}

// ---- Idempotency guard --------------------------------------------------
type dedup struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newDedup() *dedup { return &dedup{seen: map[string]bool{}} }
func (d *dedup) firstTime(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seen[id] {
		return false
	}
	d.seen[id] = true
	return true
}

func main() {
	bus := NewBus()

	// Consumer 1: inventory — idempotent (decrements stock once per event)
	invDedup := newDedup()
	stock := 100
	bus.Subscribe("OrderPlaced", func(e Event) error {
		if !invDedup.firstTime(e.ID) {
			fmt.Printf("  [inventory] duplicate %s ignored\n", e.ID)
			return nil
		}
		qty := e.Payload["qty"].(int)
		stock -= qty
		fmt.Printf("  [inventory] reserved %d (stock now %d)\n", qty, stock)
		return nil
	})

	// Consumer 2: email — independent consumer group, also idempotent
	emailDedup := newDedup()
	bus.Subscribe("OrderPlaced", func(e Event) error {
		if !emailDedup.firstTime(e.ID) {
			fmt.Printf("  [email] duplicate %s ignored\n", e.ID)
			return nil
		}
		fmt.Printf("  [email] sent confirmation for order %s\n", e.Payload["order_id"])
		return nil
	})

	fmt.Println("== Publish OrderPlaced (bus delivers each event twice) ==")
	bus.Publish(Event{ID: "evt-1", Type: "OrderPlaced",
		Payload: map[string]any{"order_id": "o1", "qty": 3}})
	bus.Publish(Event{ID: "evt-2", Type: "OrderPlaced",
		Payload: map[string]any{"order_id": "o2", "qty": 5}})

	fmt.Printf("== Final stock: %d (idempotency prevented double-decrement) ==\n", stock)
}

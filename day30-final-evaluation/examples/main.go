// Day 30 — microservices platform simulated in ONE process: a cached catalog
// (sync read, cache-aside) + an order service emitting an async OrderPlaced
// event that fans out to 3 idempotent consumers. Run: go run .
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ===== async event bus (Kafka stand-in) =====
type Event struct {
	ID      string
	Type    string
	Payload map[string]any
}
type Handler func(Event)

type Bus struct {
	mu   sync.Mutex
	subs map[string][]Handler
}

func NewBus() *Bus { return &Bus{subs: map[string][]Handler{}} }
func (b *Bus) Subscribe(t string, h Handler) {
	b.mu.Lock()
	b.subs[t] = append(b.subs[t], h)
	b.mu.Unlock()
}
func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	hs := append([]Handler(nil), b.subs[e.Type]...)
	b.mu.Unlock()
	for _, h := range hs {
		h(e)
		h(e) // at-least-once: deliver twice -> consumers must be idempotent
	}
}

// ===== sync Catalog service with cache-aside =====
type Catalog struct {
	dbCalls atomic.Int64
	mu      sync.Mutex
	cache   map[string]string
}

func NewCatalog() *Catalog { return &Catalog{cache: map[string]string{}} }
func (c *Catalog) GetProduct(_ context.Context, id string) string {
	c.mu.Lock()
	if v, ok := c.cache[id]; ok {
		c.mu.Unlock()
		return v // cache hit
	}
	c.mu.Unlock()
	c.dbCalls.Add(1)
	time.Sleep(10 * time.Millisecond) // simulate DB
	v := "Product(" + id + ")"
	c.mu.Lock()
	c.cache[id] = v
	c.mu.Unlock()
	return v
}

// ===== idempotency helper =====
type dedup struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newDedup() *dedup { return &dedup{seen: map[string]bool{}} }
func (d *dedup) first(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seen[id] {
		return false
	}
	d.seen[id] = true
	return true
}

func main() {
	ctx := context.Background()
	bus := NewBus()
	catalog := NewCatalog()

	// --- async consumers (independent groups), all idempotent ---
	pay, inv, mail := newDedup(), newDedup(), newDedup()
	bus.Subscribe("OrderPlaced", func(e Event) {
		if pay.first(e.ID) {
			fmt.Printf("  [payment]   charged order %v\n", e.Payload["order_id"])
		}
	})
	bus.Subscribe("OrderPlaced", func(e Event) {
		if inv.first(e.ID) {
			fmt.Printf("  [inventory] reserved stock for %v\n", e.Payload["order_id"])
		}
	})
	bus.Subscribe("OrderPlaced", func(e Event) {
		if mail.first(e.ID) {
			fmt.Printf("  [email]     sent receipt for %v\n", e.Payload["order_id"])
		}
	})

	fmt.Println("== SYNC: gateway reads catalog (cache-aside) ==")
	for i := 0; i < 3; i++ {
		fmt.Printf("  catalog.GetProduct(p1) -> %s\n", catalog.GetProduct(ctx, "p1"))
	}
	fmt.Printf("  catalog DB calls: %d (cached after first)\n", catalog.dbCalls.Load())

	fmt.Println("== ASYNC: order service emits OrderPlaced (fans out) ==")
	bus.Publish(Event{ID: "evt-1001", Type: "OrderPlaced",
		Payload: map[string]any{"order_id": "o-1001"}})

	fmt.Println("== Platform demo complete: sync query + async fan-out, idempotent ==")
}

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Job int

// Queue is "thread-safe" (mutex-guarded) but UNBOUNDED: the backing slice grows
// without limit, so Submit never applies backpressure.
type Queue struct {
	mu      sync.Mutex
	jobs    []Job
	maxSeen int64 // peak backlog, for the harness
	done    int64 // processed count
}

func NewQueue(workers int) *Queue {
	q := &Queue{}
	for i := 0; i < workers; i++ {
		go q.worker()
	}
	return q
}

// BUG: append grows the slice forever. A fast producer is never slowed down.
func (q *Queue) Submit(j Job) {
	q.mu.Lock()
	q.jobs = append(q.jobs, j)
	if n := int64(len(q.jobs)); n > q.maxSeen {
		q.maxSeen = n
	}
	q.mu.Unlock()
}

func (q *Queue) worker() {
	for {
		q.mu.Lock()
		if len(q.jobs) == 0 {
			q.mu.Unlock()
			time.Sleep(time.Millisecond)
			continue
		}
		q.jobs = q.jobs[1:]
		q.mu.Unlock()

		time.Sleep(time.Millisecond) // slow processing
		atomic.AddInt64(&q.done, 1)
	}
}

func main() {
	const total = 50000
	q := NewQueue(4) // few, slow workers

	start := time.Now()
	for i := 0; i < total; i++ {
		q.Submit(Job(i)) // fast producer; never blocks
	}
	submitElapsed := time.Since(start)

	// wait for drain
	for atomic.LoadInt64(&q.done) < total {
		time.Sleep(5 * time.Millisecond)
	}

	q.mu.Lock()
	peak := q.maxSeen
	q.mu.Unlock()

	fmt.Printf("submitted %d jobs in %v (producer never blocked)\n", total, submitElapsed.Round(time.Millisecond))
	fmt.Printf("peak backlog: %d (want it bounded, near a small capacity)\n", peak)
}

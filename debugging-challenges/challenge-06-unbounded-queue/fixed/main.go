package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Job int

// Queue uses a BUFFERED channel: capacity is a hard ceiling, and a send on a
// full channel blocks — that block is the backpressure.
type Queue struct {
	jobs     chan Job
	maxSeen  int64 // peak backlog (channel length)
	done     int64
	capacity int
	wg       sync.WaitGroup
}

func NewQueue(capacity, workers int) *Queue {
	q := &Queue{
		jobs:     make(chan Job, capacity),
		capacity: capacity,
	}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
	return q
}

// Submit blocks when the buffer is full -> the producer is throttled to worker
// speed. Memory is capped at `capacity` queued jobs.
func (q *Queue) Submit(j Job) {
	q.jobs <- j
	if n := int64(len(q.jobs)); n > atomic.LoadInt64(&q.maxSeen) {
		atomic.StoreInt64(&q.maxSeen, n)
	}
}

func (q *Queue) Close() {
	close(q.jobs)
	q.wg.Wait()
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for range q.jobs {
		time.Sleep(time.Millisecond) // slow processing
		atomic.AddInt64(&q.done, 1)
	}
}

func main() {
	const (
		total    = 50000
		capacity = 256
	)
	q := NewQueue(capacity, 4)

	start := time.Now()
	for i := 0; i < total; i++ {
		q.Submit(Job(i)) // blocks once the buffer fills
	}
	submitElapsed := time.Since(start)

	q.Close() // drain remaining jobs and stop workers

	fmt.Printf("submitted %d jobs in %v (producer throttled by backpressure)\n", total, submitElapsed.Round(time.Millisecond))
	fmt.Printf("queue capacity: %d\n", capacity)
	fmt.Printf("peak backlog: %d (bounded by capacity)\n", atomic.LoadInt64(&q.maxSeen))
	fmt.Printf("processed: %d\n", atomic.LoadInt64(&q.done))
}

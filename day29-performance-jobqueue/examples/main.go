// Day 29 walkthrough — a job queue with bounded workers, retries with
// exponential backoff + jitter, and a dead-letter queue. Run: go run .
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID       string
	Attempts int
	Max      int
	run      func() error // the work; returns error on failure
}

type Queue struct {
	jobs        chan Job
	wg          sync.WaitGroup
	mu          sync.Mutex
	succeeded   []string
	deadLetters []string
	rng         uint64 // simple deterministic PRNG state for jitter
}

func NewQueue(buffer int) *Queue {
	return &Queue{jobs: make(chan Job, buffer), rng: 12345}
}

// deterministic jitter in [0, n) ms — avoids Math.random (unavailable) and
// keeps the demo reproducible. xorshift.
func (q *Queue) jitter(n int) time.Duration {
	q.mu.Lock()
	x := q.rng
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	q.rng = x
	q.mu.Unlock()
	if n <= 0 {
		return 0
	}
	return time.Duration(int(x%uint64(n))) * time.Millisecond
}

func (q *Queue) Submit(j Job) {
	q.wg.Add(1)
	q.jobs <- j
}

func (q *Queue) backoff(attempt int) time.Duration {
	base := 5 * time.Millisecond
	d := base << attempt // exponential: base * 2^attempt
	return d + q.jitter(5)
}

func (q *Queue) worker(id int) {
	for j := range q.jobs {
		err := j.run()
		if err == nil {
			q.mu.Lock()
			q.succeeded = append(q.succeeded, j.ID)
			q.mu.Unlock()
			fmt.Printf("  [w%d] %s ok (attempt %d)\n", id, j.ID, j.Attempts+1)
			q.wg.Done()
			continue
		}
		j.Attempts++
		if j.Attempts >= j.Max {
			q.mu.Lock()
			q.deadLetters = append(q.deadLetters, j.ID)
			q.mu.Unlock()
			fmt.Printf("  [w%d] %s DEAD-LETTERED after %d attempts: %v\n", id, j.ID, j.Attempts, err)
			q.wg.Done()
			continue
		}
		delay := q.backoff(j.Attempts)
		fmt.Printf("  [w%d] %s failed (attempt %d), retry in %v: %v\n", id, j.ID, j.Attempts, delay, err)
		// requeue after backoff without blocking the worker
		go func(j Job, d time.Duration) {
			time.Sleep(d)
			q.jobs <- j
		}(j, delay)
	}
}

func (q *Queue) Start(workers int) {
	for i := 1; i <= workers; i++ {
		go q.worker(i)
	}
}
func (q *Queue) Wait() {
	q.wg.Wait()
	close(q.jobs)
}

func main() {
	q := NewQueue(64)
	q.Start(3)

	// Job A: flaky — fails first 2 attempts, then succeeds (retry recovers it)
	var aTries int
	q.Submit(Job{ID: "job-A", Max: 5, run: func() error {
		aTries++
		if aTries < 3 {
			return fmt.Errorf("transient error")
		}
		return nil
	}})

	// Job B: always succeeds first try
	q.Submit(Job{ID: "job-B", Max: 5, run: func() error { return nil }})

	// Job C: permanently broken -> exhausts retries -> dead-letter
	q.Submit(Job{ID: "job-C", Max: 3, run: func() error {
		return fmt.Errorf("permanent failure")
	}})

	fmt.Println("== processing jobs (workers=3) ==")
	q.Wait()

	fmt.Printf("\n== Results ==\n  succeeded: %v\n  dead-letters: %v\n", q.succeeded, q.deadLetters)
}

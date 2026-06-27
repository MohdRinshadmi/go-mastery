// Day 29 — YOUR job-queue project. Run: go run .
package main

import "fmt"

// =====================================================================
// Build a job queue with:
//  - a bounded worker pool (e.g. 3 workers) pulling from a jobs channel
//  - retries with EXPONENTIAL BACKOFF + JITTER up to job.Max attempts
//  - a DEAD-LETTER slice for jobs that exhaust retries
//  - track succeeded job IDs
//
// Use a sync.WaitGroup that completes when every job has reached a terminal
// state (succeeded or dead-lettered) — note a retried job is NOT done yet.
//
// NOTE: Math.random / time.Now-based seeding are fine in YOUR local run
// (this isn't a workflow). A simple xorshift for jitter also works.
//
// Demonstrate with:
//   job-A: fails twice then succeeds (retry recovers)
//   job-B: succeeds immediately
//   job-C: always fails -> dead-letter after Max attempts
// =====================================================================

type Job struct {
	ID       string
	Attempts int
	Max      int
	run      func() error
}

// TODO: type Queue struct { jobs chan Job; wg; succeeded; deadLetters; ... }
// TODO: NewQueue, Submit, worker (with backoff+requeue), Start, Wait

func main() {
	fmt.Println("TODO: implement the job queue; see lesson + ../solutions")
}

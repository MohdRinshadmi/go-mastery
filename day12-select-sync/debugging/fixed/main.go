// Day 12 debugging — FIXED.
//
// Same "run N tasks per round, wait for all" scenario, done correctly.
//
// Fix: call wg.Add(1) in the loop, BEFORE `go`. The counter is incremented
// before any goroutine can run, so wg.Wait() cannot return until every task has
// called Done(). Every round completes all `tasks`, and it's clean under -race.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func runRound(tasks int) int64 {
	var wg sync.WaitGroup
	var completed atomic.Int64

	for i := 0; i < tasks; i++ {
		wg.Add(1) // FIX: Add before launching the goroutine
		go func(id int) {
			defer wg.Done()
			completed.Add(1)
		}(i)
	}

	wg.Wait() // guaranteed to block until all `tasks` goroutines call Done
	return completed.Load()
}

func main() {
	const tasks = 50
	const rounds = 8

	allGood := true
	for r := 0; r < rounds; r++ {
		if got := runRound(tasks); got != int64(tasks) {
			allGood = false
			fmt.Printf("round %d: completed %d/%d\n", r, got, tasks)
		}
	}

	if allGood {
		fmt.Printf("all %d rounds completed every one of %d tasks before Wait returned\n", rounds, tasks)
	}
}

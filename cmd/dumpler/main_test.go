package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type Job struct {
	ID int
}

func worker(id int, jobs <-chan Job, done <-chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case job, ok := <-jobs:
			if ok {
				fmt.Printf("Worker %d processing job %d\n", id, job.ID)
				time.Sleep(1 * time.Second) // simulate job processing time
			} else {
				wg.Done()
				return
			}
		case <-done:
			wg.Done()
			return
		}
	}
}

func TestJobProcessing(t *testing.T) {
	testCases := []struct {
		name       string
		numWorkers int
		numJobs    int
	}{
		{"1 worker, 5 jobs", 1, 5},
		{"2 workers, 5 jobs", 2, 5},
		{"3 workers, 5 jobs", 3, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jobs := make(chan Job)
			done := make(chan bool)
			var wg sync.WaitGroup

			for i := 1; i <= tc.numWorkers; i++ {
				wg.Add(1)
				go worker(i, jobs, done, &wg)
			}

			go func() {
				for i := 1; i <= tc.numJobs; i++ {
					jobs <- Job{ID: i}
				}
				close(jobs)
				done <- true
			}()

			wg.Wait()
		})
	}
}

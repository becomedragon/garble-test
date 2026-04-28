// Package worker provides goroutine-based task processing with channels and select.
package worker

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/becomedragon/garble-test/models"
)

const (
	defaultWorkerCount = 3
	jobBufferSize      = 16
	resultBufferSize   = 16
)

// Job carries the input for a worker.
type Job struct {
	ID    int
	Input string
}

// Result carries the output of a processed job.
type Result struct {
	JobID  int
	Output string
	Err    error
}

// Pool manages a fixed number of worker goroutines.
type Pool struct {
	size    int
	jobs    chan Job
	results chan Result
	quit    chan struct{}
	wg      sync.WaitGroup
}

// NewPool creates a Pool with the given number of workers.
func NewPool(size int) *Pool {
	if size <= 0 {
		size = defaultWorkerCount
	}
	return &Pool{
		size:    size,
		jobs:    make(chan Job, jobBufferSize),
		results: make(chan Result, resultBufferSize),
		quit:    make(chan struct{}),
	}
}

// Start launches the worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.runWorker(i)
	}
}

// Submit sends a job to the pool.  It blocks if the job buffer is full.
func (p *Pool) Submit(j Job) {
	p.jobs <- j
}

// Results returns the channel on which results are delivered.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop signals all workers to finish and waits for them.
func (p *Pool) Stop() {
	close(p.quit)
	p.wg.Wait()
	close(p.results)
}

// CloseWhenDone waits for all workers to finish and then closes the results
// channel so that callers ranging over Results() will unblock.
func (p *Pool) CloseWhenDone() {
	p.wg.Wait()
	close(p.results)
}

// runWorker is the main loop for a single worker goroutine.
func (p *Pool) runWorker(id int) {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			res := processJob(id, job)
			select {
			case p.results <- res:
			case <-p.quit:
				return
			}
		case <-p.quit:
			return
		}
	}
}

// processJob applies the business logic to a single job.
func processJob(workerID int, j Job) Result {
	// Simulate variable processing time.
	delay := time.Duration(rand.Intn(50)) * time.Millisecond
	time.Sleep(delay)

	var output string
	switch {
	case len(j.Input) == 0:
		output = fmt.Sprintf("worker%d: empty input for job %d", workerID, j.ID)
	case len(j.Input) < 5:
		output = fmt.Sprintf("worker%d: short input %q", workerID, j.Input)
	default:
		output = fmt.Sprintf("worker%d: processed job %d → %q", workerID, j.ID, j.Input)
	}
	return Result{JobID: j.ID, Output: output}
}

// RunTaskPipeline processes a slice of models.Task objects concurrently and
// returns a models.Report summarising the results.
func RunTaskPipeline(tasks []*models.Task, workerCount int) *models.Report {
	pool := NewPool(workerCount)
	pool.Start()

	// Feed jobs in a goroutine so Submit does not block the collector.
	go func() {
		for _, t := range tasks {
			pool.Submit(Job{ID: t.ID, Input: t.Title})
		}
		// Closing jobs signals workers that no more work is coming.
		close(pool.jobs)
	}()

	// Once all workers are done, close the results channel so the range loop
	// below will terminate naturally (no separate Stop call needed here).
	go pool.CloseWhenDone()

	report := models.NewReport("pipeline")

	// Collect results; the loop exits when pool.CloseWhenDone closes the channel.
	// TODO: in a production implementation, map res.JobID back to the task and
	// use res.Output / res.Err to update each task's state individually.
	for res := range pool.Results() {
		_ = res
	}

	for _, t := range tasks {
		t.Complete()
		report.AddTask(t)
	}

	return report
}

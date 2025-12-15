package patron

import (
	"context"
	"sync"
)

type WorkerOrchestrator interface {
	GetQueueLength() int
	AddJobToQueue(job *Job)
	Start(context.Context) []WorkerResult
}

type Config struct {
	WorkerCount int
	WorkerFunc  func(job *Job) error
}

type workerOrchestrator struct {
	workerCount int
	workerFunc  func(job *Job) error
	jobs        []*Job
	mu          sync.Mutex
}

// New creates a new worker orchestrator.
func New(conf Config) WorkerOrchestrator {
	return &workerOrchestrator{
		workerCount: conf.WorkerCount,
		workerFunc:  conf.WorkerFunc,
		jobs:        make([]*Job, 0),
	}
}

// AddJobToQueue adds a new job to the queue.
func (wo *workerOrchestrator) AddJobToQueue(job *Job) {
	wo.mu.Lock()
	defer wo.mu.Unlock()
	wo.jobs = append(wo.jobs, job)
}

// GetQueueLength returns the length of the job queue.
func (wo *workerOrchestrator) GetQueueLength() int {
	wo.mu.Lock()
	defer wo.mu.Unlock()
	return len(wo.jobs)
}

// Start starts all the jobs in the queue using a worker pool.
func (wo *workerOrchestrator) Start(ctx context.Context) []WorkerResult {
	// If no jobs, return empty
	wo.mu.Lock()
	jobCount := len(wo.jobs)
	jobs := wo.jobs
	wo.mu.Unlock()

	if jobCount == 0 {
		return nil
	}

	jobsCh := make(chan *Job, jobCount)
	resultsCh := make(chan WorkerResult, jobCount)
	var wg sync.WaitGroup

	// Push all jobs to channel
	for _, job := range jobs {
		jobsCh <- job
	}
	close(jobsCh)

	// Start workers
	// If jobCount < workerCount, we only need jobCount workers (optimization)
	workersToStart := wo.workerCount
	if jobsChLen := len(jobs); jobsChLen < workersToStart {
		// workersToStart = jobsChLen // Optional optimization, but maybe overhead to logic
	}

	for i := 0; i < workersToStart; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobsCh {
				select {
				case <-ctx.Done():
					return
				default:
					err := wo.workerFunc(job)
					resultsCh <- WorkerResult{
						WorkerID: workerID,
						JobID:    job.ID,
						Error:    err,
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(resultsCh)

	var results []WorkerResult
	for res := range resultsCh {
		results = append(results, res)
	}

	// Clear the queue after processing?
	// The original implementation seemed to consume them.
	wo.mu.Lock()
	wo.jobs = nil
	wo.mu.Unlock()

	return results
}

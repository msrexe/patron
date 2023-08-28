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
	jobQueue      []*Job
	jobQueueIndex int

	workers []*worker
}

// Newcreates a new worker orchestrator.
func New(conf Config) WorkerOrchestrator {
	return &workerOrchestrator{
		workers: newWorkerArray(conf.WorkerFunc, conf.WorkerCount),
	}
}

// AddJobToQueue adds a new job to the queue.
func (wo *workerOrchestrator) AddJobToQueue(job *Job) {
	wo.jobQueue = append(wo.jobQueue, job)
}

// GetQueueLength returns the length of the job queue.
func (wo *workerOrchestrator) GetQueueLength() int {
	return len(wo.jobQueue) - wo.jobQueueIndex
}

// TODO: Implement this method.
// func (wo *workerOrchestrator) StartAsAsync(ctx context.Context, workerResultCh chan *WorkerResult)

// Start starts all the jobs in the queue.
func (wo *workerOrchestrator) Start(ctx context.Context) []WorkerResult {
	var wg sync.WaitGroup
	var workerResult *WorkerResult
	var workerResults []WorkerResult

	totalJobCount := wo.GetQueueLength()
	resultCh := make(chan *WorkerResult, totalJobCount)

	for i := 0; i < totalJobCount; {
		if wo.findAvailableWorker() != nil {
			job := wo.consumeJobFromQueue()
			wrk := wo.findAvailableWorker()
			wrk.SetJob(job)

			wg.Add(1)
			go func(wrk *worker, resultCh chan *WorkerResult) {
				defer wrk.FinalizeJob()
				defer wg.Done()

				resultCh <- &WorkerResult{
					WorkerID: wrk.GetID(),
					JobID:    wrk.GetJob().ID,
					Error:    wrk.Work(),
				}
			}(wrk, resultCh)

			i++
		} else {
			// If no available worker, waits for a worker to be free
			workerResult = <-resultCh
			workerResults = append(workerResults, *workerResult)
		}
	}

	wg.Wait()
	close(resultCh)

	for result := range resultCh {
		workerResults = append(workerResults, *result)
	}

	return workerResults
}

func (wo *workerOrchestrator) consumeJobFromQueue() *Job {
	if wo.GetQueueLength() > 0 && wo.jobQueueIndex < len(wo.jobQueue) {
		job := wo.jobQueue[wo.jobQueueIndex]
		wo.jobQueueIndex++
		return job
	}

	wo.jobQueueIndex = 0
	return nil
}

func (wo *workerOrchestrator) findAvailableWorker() *worker {
	for _, worker := range wo.workers {
		if !worker.IsBusy() {
			return worker
		}
	}

	return nil
}

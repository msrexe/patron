package patron

import (
	"context"
	"sync"
)

type WorkerOrchestrator interface {
	GetQueueLength() int
	AddJobToQueue(job *Job)
	Start(context.Context) []WorkerResult
	StartAsAsync(context.Context, chan *WorkerResult)
}

type Config struct {
	WorkerCount int
	WorkerFunc  func(job *Job) error
}

type workerOrchestrator struct {
	jobQueue []*Job
	workers  []Worker
}

// NewWorkerOrchestrator creates a new worker orchestrator.
func NewWorkerOrchestrator(conf Config) WorkerOrchestrator {
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
	return len(wo.jobQueue)
}

// TODO: Implement this method.
func (wo *workerOrchestrator) StartAsAsync(ctx context.Context, workerResultCh chan *WorkerResult) {
	var wg sync.WaitGroup

	for job := wo.consumeJobFromQueue(); job != nil; job = wo.consumeJobFromQueue() {
		for worker := wo.findAvailableWorker(); worker != nil; worker = wo.findAvailableWorker() {
			worker.SetJob(job)

			wg.Add(1)
			go func(worker Worker) {
				defer worker.FinalizeJob()
				defer wg.Done()

				workerResultCh <- &WorkerResult{
					WorkerID: worker.GetID(),
					JobID:    worker.GetJob().ID,
					Error:    worker.Work(),
				}
			}(worker)
		}
	}

	wg.Wait()
}

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
			worker := wo.findAvailableWorker()
			worker.SetJob(job)

			wg.Add(1)
			go func(worker Worker, resultCh chan *WorkerResult) {
				defer worker.FinalizeJob()
				defer wg.Done()

				resultCh <- &WorkerResult{
					WorkerID: worker.GetID(),
					JobID:    worker.GetJob().ID,
					Error:    worker.Work(),
				}
			}(worker, resultCh)

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
	if len(wo.jobQueue) > 0 {
		job := wo.jobQueue[0]
		wo.jobQueue[0] = nil
		wo.jobQueue = wo.jobQueue[1:]
		return job
	}

	return nil
}

func (wo *workerOrchestrator) findAvailableWorker() Worker {
	for _, worker := range wo.workers {
		if !worker.IsBusy() {
			return worker
		}
	}

	return nil
}

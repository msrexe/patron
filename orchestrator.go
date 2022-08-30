package patron

import (
	"context"
	"sync"
)

type WorkerOrchestrator interface {
	GetQueueLength() int
	AddJobToQueue(job *Job)
	Start(context.Context) []*WorkerResult
	StartAsAsync(context.Context, chan *WorkerResult)
}

type workerOrchestrator struct {
	jobQueue []*Job
	workers  []Worker
}

func NewWorkerOrchestrator(workers []Worker) WorkerOrchestrator {
	return &workerOrchestrator{
		workers: workers,
	}
}

func (wo *workerOrchestrator) AddJobToQueue(job *Job) {
	wo.jobQueue = append(wo.jobQueue, job)
}

func (wo *workerOrchestrator) GetQueueLength() int {
	return len(wo.jobQueue)
}

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

func (wo *workerOrchestrator) Start(ctx context.Context) []*WorkerResult {
	var wg sync.WaitGroup
	var workerResults []*WorkerResult

	totalJobCount := wo.GetQueueLength()

	for i := 0; i < totalJobCount && wo.findAvailableWorker() != nil; i++ {
		job := wo.consumeJobFromQueue()
		worker := wo.findAvailableWorker()
		worker.SetJob(job)

		wg.Add(1)
		go func(worker Worker) {
			defer worker.FinalizeJob()
			defer wg.Done()

			workerResults = append(workerResults, &WorkerResult{
				WorkerID: worker.GetID(),
				JobID:    worker.GetJob().ID,
				Error:    worker.Work(),
			})
		}(worker)
	}

	wg.Wait()

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

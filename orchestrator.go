package patron

import (
	"context"
	"sync"
)

const (
	MAX_WORKER_COUNT     = 255
	DEFAULT_WORKER_COUNT = 10
	MAX_JOB_QUEUE_SIZE   = 32767
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
	workerResults []WorkerResult

	workers []*worker
}

func New(conf Config) WorkerOrchestrator {
	if conf.WorkerCount <= 0 {
		conf.WorkerCount = DEFAULT_WORKER_COUNT
	}

	return &workerOrchestrator{
		jobQueue:      make([]*Job, 0, MAX_JOB_QUEUE_SIZE),
		workerResults: make([]WorkerResult, 0, MAX_JOB_QUEUE_SIZE),
		workers:       newWorkerArray(conf.WorkerFunc, conf.WorkerCount),
	}
}

func (wo *workerOrchestrator) AddJobToQueue(job *Job) {
	wo.jobQueue = append(wo.jobQueue, job)
}

func (wo *workerOrchestrator) GetQueueLength() int {
	return len(wo.jobQueue) - wo.jobQueueIndex
}

func (wo *workerOrchestrator) Start(ctx context.Context) []WorkerResult {
	var wg sync.WaitGroup

	totalJobCount := wo.GetQueueLength()
	jobCh := make(chan *Job, totalJobCount)

	wo.workerResults = wo.workerResults[:0]

	for _, wrk := range wo.workers {
		wg.Add(1)
		go func(wrk *worker) {
			defer wrk.FinalizeJob()
			defer wg.Done()

			for job := range jobCh {
				wrk.SetJob(job)
				wo.workerResults = append(wo.workerResults, WorkerResult{
					WorkerID: wrk.GetID(),
					JobID:    wrk.GetJob().ID,
					Error:    wrk.Work(),
				})
			}
		}(wrk)
	}

	for i := 0; i < totalJobCount; i++ {
		jobCh <- wo.consumeJobFromQueue()
	}

	close(jobCh)
	wg.Wait()

	return wo.workerResults
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

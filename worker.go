package patron

type worker struct {
	id         int
	job        *Job
	workerFunc func(job *Job) error
}

func newWorkerArray(workerFunc func(job *Job) error, workerCount int) []*worker {
	switch {
	case workerCount <= 0:
		workerCount = DEFAULT_WORKER_COUNT
	case workerCount > MAX_WORKER_COUNT:
		workerCount = MAX_WORKER_COUNT
	}

	workers := make([]*worker, workerCount, MAX_WORKER_COUNT)

	for i := 0; i < workerCount; i++ {
		workers[i] = newWorker(i, workerFunc)
	}

	return workers
}

func newWorker(id int, workerFunc func(job *Job) error) *worker {
	return &worker{
		id:         id,
		workerFunc: workerFunc,
	}
}

func (w *worker) GetID() int {
	return w.id
}

func (w *worker) GetJob() *Job {
	return w.job
}

func (w *worker) SetJob(job *Job) {
	w.job = job
}

func (w *worker) FinalizeJob() {
	w.job = nil
}

func (w *worker) IsBusy() bool {
	return w.job != nil
}

func (w *worker) Work() error {
	return w.workerFunc(w.job)
}

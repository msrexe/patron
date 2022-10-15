package patron

type Worker interface {
	GetID() int
	GetJob() *Job
	SetJob(*Job)
	FinalizeJob()
	IsBusy() bool
	Work() error
}

type worker struct {
	id         int
	job        *Job
	workerFunc *func(job *Job) error
}

func newWorkerArray(workerFunc func(job *Job) error, workerCount int) []Worker {
	workers := make([]Worker, workerCount)

	for i := 0; i < workerCount; i++ {
		workers[i] = newWorker(i, &workerFunc)
	}

	return workers
}

func newWorker(id int, workerFunc *func(job *Job) error) *worker {
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
	fun := *w.workerFunc
	return fun(w.job)
}

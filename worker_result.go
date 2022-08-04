package patron

type WorkerResult struct {
	WorkerID int
	JobID    int
	Error    error
}

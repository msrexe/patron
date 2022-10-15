package main

import (
	"context"
	"fmt"
	"time"

	"github.com/msrexe/patron"
)

func main() {
	ctx := context.Background()

	// Start the orchestrator.
	workerOrchestrator := patron.NewWorkerOrchestrator(
		patron.Config{
			WorkerCount: 5,
			WorkerFunc:  workerFunction,
		},
	)

	// Add jobs to the orchestrator.
	for i := 0; i < 10; i++ {
		workerOrchestrator.AddJobToQueue(&patron.Job{
			ID:      i,
			Context: ctx,
			Payload: map[string]interface{}{
				"name": "HTTP Request #" + fmt.Sprintf("%d", i),
			},
		})
	}

	workerResults := workerOrchestrator.Start(context.Background())

	for _, workerResult := range workerResults {
		if workerResult.Error != nil {
			fmt.Printf("Worker %d failed with error: %s\n", workerResult.WorkerID, workerResult.Error.Error())
		} else {
			fmt.Printf("Worker %d finished job %d\n", workerResult.WorkerID, workerResult.JobID)
		}
	}
}

func workerFunction(job *patron.Job) error {
	time.Sleep(1 * time.Second)
	payloadName, err := job.GetPayload("name")
	if err != nil {
		return err
	}

	fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, payloadName)

	return nil
}

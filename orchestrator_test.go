package patron

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func setupTestOrchestrator() WorkerOrchestrator {
	return New(
		Config{
			WorkerCount: 5,
			WorkerFunc: func(job *Job) error {
				time.Sleep(1 * time.Second)
				payloadName, err := job.Get("name")
				if err != nil {
					return err
				}

				fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, payloadName)

				return nil
			},
		},
	)
}

func TestNewWorkerOrchestrator(t *testing.T) {
	orch := New(Config{WorkerCount: 5, WorkerFunc: nil})
	if orch == nil {
		t.Error("Expected new worker orchestrator to be not nil")
	}
}

func TestAddJobToQueue(t *testing.T) {
	workerOrchestrator := setupTestOrchestrator()

	workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	if len := workerOrchestrator.GetQueueLength(); len != 1 {
		t.Errorf("Expected queue length 1, got %d", len)
	}
}

func TestStartAllJobsSuccess(t *testing.T) {
	workerOrchestrator := setupTestOrchestrator()

	workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test",
		},
	})
	workerOrchestrator.AddJobToQueue(&Job{
		ID:      11,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test2",
		},
	})

	results := workerOrchestrator.Start(context.Background())

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Errorf("Expected no error for job 1, got %v", results[0].Error)
	}
	if results[1].Error != nil {
		t.Errorf("Expected no error for job 2, got %v", results[1].Error)
	}
}

func TestStartOneJobFailure(t *testing.T) {
	workerOrchestrator := setupTestOrchestrator()

	workerOrchestrator.AddJobToQueue(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/test",
		},
	})
	workerOrchestrator.AddJobToQueue(&Job{
		ID:      11,
		Context: context.Background(),
		Payload: map[string]interface{}{},
	})

	results := workerOrchestrator.Start(context.Background())

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.JobID == 11 {
			if !errors.Is(result.Error, ErrJobPayloadNotFound) {
				t.Errorf("Expected ErrJobPayloadNotFound for job 11, got %v", result.Error)
			}
		} else {
			if result.Error != nil {
				t.Errorf("Expected no error for job %d, got %v", result.JobID, result.Error)
			}
		}
	}
}

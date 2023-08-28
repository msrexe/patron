package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/msrexe/patron"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	wg.Add(1)
	fmt.Print("Starting patron...\n")
	time.Sleep(5 * time.Second)
	// Create the orchestrator.
	workerOrchestrator := patron.New(
		patron.Config{
			WorkerCount: 5,
			WorkerFunc:  workerFunction,
		},
	)

	// Add jobs to the orchestrator.
	for i := 0; i < 10_000; i++ {
		workerOrchestrator.AddJobToQueue(&patron.Job{
			ID:      i,
			Context: context.Background(),
			Payload: map[string]interface{}{
				"name": fmt.Sprintf("test_%d", i),
			},
		})
	}

	// Start timer
	start := time.Now()
	workerResults := workerOrchestrator.Start(context.Background())
	elapsed := time.Since(start)

	for _, workerResult := range workerResults {
		if workerResult.Error != nil {
			fmt.Printf("Worker %d failed with error: %s\n", workerResult.WorkerID, workerResult.Error.Error())
		} else {
			fmt.Printf("Worker %d finished job %d\n", workerResult.WorkerID, workerResult.JobID)
		}
	}

	// Time elapsed: 472.747625ms
	fmt.Printf("Time elapsed: %s\n", elapsed)

	wg.Wait()
}

func workerFunction(job *patron.Job) error {
	fileName, err := job.GetPayload("name")
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", "_test", fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d", job.ID))
	if err != nil {
		return err
	}

	fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, fileName)

	return nil
}

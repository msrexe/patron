package benchmark

import (
	"context"
	"math"
	"testing"

	"github.com/msrexe/patron"
)

var (
	results        = make([]float64, 10_000)
	workerFunction = func(job *patron.Job) error {
		num, err := job.GetPayload("num")
		if err != nil {
			return err
		}

		result := math.Sqrt(num.(float64))

		found := false
		for _, res := range results {
			if res == result {
				found = true
				break
			}
		}

		if !found {
			results = append(results, result)
		}
		return nil
	}
)

// BenchmarkPatron-8   	  355014	      5317 ns/op	     735 B/op	       7 allocs/op
// BenchmarkPatron-8   	  390490	      5416 ns/op	     296 B/op	       3 allocs/op
func BenchmarkPatron(b *testing.B) {
	orc := patron.New(patron.Config{
		WorkerCount: 5,
		WorkerFunc:  workerFunction,
	})

	for i := 0; i < b.N; i++ {
		orc.AddJobToQueue(&patron.Job{
			ID:      i,
			Context: nil,
			Payload: map[string]interface{}{
				"num": float64(i * 3),
			},
		})
	}

	b.ResetTimer()
	_ = orc.Start(context.Background())
}

// BenchmarkClassicPoolImpl-8   	20819346	        70.12 ns/op	      13 B/op	       0 allocs/op
// BenchmarkClassicPoolImpl-8   	21875970	        61.93 ns/op	       7 B/op	       0 allocs/op
func BenchmarkClassicPoolImpl(b *testing.B) {
	numJobs := b.N

	jobs := make(chan int, numJobs)

	for w := 1; w <= 5; w++ {
		go worker(w, jobs)
	}

	b.ResetTimer()
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}

	close(jobs)
}

func worker(id int, jobs <-chan int) {
	for j := range jobs {
		workerFunction(&patron.Job{
			ID:      j,
			Context: nil,
			Payload: map[string]interface{}{
				"num": float64(j * 3),
			},
		})
	}
}

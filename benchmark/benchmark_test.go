package benchmark

import (
	"context"
	"sync"
	"testing"

	"github.com/gammazero/workerpool"
	"github.com/msrexe/patron"
	"github.com/panjf2000/ants/v2"
	"github.com/sourcegraph/conc/iter"
	"golang.org/x/sync/errgroup"
)

// Benchmark for WorkerOrchestrator (Structural Pool)
// This strictly tests Patron's internal Orchestrator performance.
func BenchmarkWorkerOrchestrator(b *testing.B) {
	workerFunc := func(job *patron.Job) error {
		_ = job.ID * 2
		return nil
	}

	b.Run("100Jobs_10Workers", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			orch := patron.New(patron.Config{WorkerCount: 10, WorkerFunc: workerFunc})
			for i := 0; i < 100; i++ {
				orch.AddJobToQueue(&patron.Job{ID: i})
			}
			orch.Start(context.Background())
		}
	})

	b.Run("1000Jobs_100Workers", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			orch := patron.New(patron.Config{WorkerCount: 100, WorkerFunc: workerFunc})
			for i := 0; i < 1000; i++ {
				orch.AddJobToQueue(&patron.Job{ID: i})
			}
			orch.Start(context.Background())
		}
	})

	b.Run("10000Jobs_500Workers", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			orch := patron.New(patron.Config{WorkerCount: 500, WorkerFunc: workerFunc})
			for i := 0; i < 10000; i++ {
				orch.AddJobToQueue(&patron.Job{ID: i})
			}
			orch.Start(context.Background())
		}
	})
}

// Benchmark comparing different Pool implementations
func BenchmarkPoolComparison(b *testing.B) {
	// Common workload
	work := func(id int) {
		_ = id * 2
	}

	workerFunc := func(job *patron.Job) error {
		work(job.ID)
		return nil
	}

	jobCount := 10000
	workerCount := 100

	// 1. Patron Worker Orchestrator
	b.Run("Patron_Orchestrator", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			orch := patron.New(patron.Config{WorkerCount: workerCount, WorkerFunc: workerFunc})
			for i := 0; i < jobCount; i++ {
				orch.AddJobToQueue(&patron.Job{ID: i})
			}
			orch.Start(context.Background())
		}
	})

	// 2. Panjf2000 / Ants
	b.Run("Ants_Pool", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var wg sync.WaitGroup
			p, _ := ants.NewPoolWithFunc(workerCount, func(i interface{}) {
				work(i.(int))
				wg.Done()
			})
			for i := 0; i < jobCount; i++ {
				wg.Add(1)
				_ = p.Invoke(i)
			}
			wg.Wait()
			p.Release()
		}
	})

	// 3. Gammazero / WorkerPool
	b.Run("Gammazero_WorkerPool", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			wp := workerpool.New(workerCount)
			for i := 0; i < jobCount; i++ {
				i := i
				wp.Submit(func() {
					work(i)
				})
			}
			wp.StopWait()
		}
	})
}

// Benchmark for Iteration Helpers (ForEach)
func BenchmarkForEachComparison(b *testing.B) {
	count := 10000
	items := make([]int, count)
	for i := range items {
		items[i] = i
	}

	// Workload to simulate some actual processing
	work := func(i int) {
		_ = i * 2
	}

	// 1. Patron ForEach
	b.Run("Patron_ForEach", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			patron.ForEach(items, func(item int) {
				work(item)
			})
		}
	})

	// 2. Conc ForEach (Sourcegraph)
	b.Run("Conc_ForEach", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			iter.ForEach(items, func(item *int) {
				work(*item)
			})
		}
	})

	// 3. ErrGroup (golang.org/x/sync)
	b.Run("Sync_ErrGroup", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var g errgroup.Group
			// ErrGroup doesn't limit concurrency by default unless SetLimit is used (Go 1.19+)
			// To match ForEach (GOMAXPROCS), we implicitly trust the scheduler or we should set limit?
			// ForEach usually runs at GOMAXPROCS.
			// Let's us SetLimit if available or just run.
			// Given simple ForEach semantic comparison, we just launch.
			// But creating 10000 goroutines is unfair if ForEach uses a pool.
			// Patron and Conc ForEach utilize a limited number of workers (GOMAXPROCS usually).
			// So for ErrGroup to be fair we should limit it.
			g.SetLimit(10) // Approx GOMAXPROCS for M1 is usually 8 or 10.

			for _, item := range items {
				item := item
				g.Go(func() error {
					work(item)
					return nil
				})
			}
			_ = g.Wait()
		}
	})

	// 4. Manual WaitGroup (Idiomatic Go - Unbounded)
	b.Run("Manual_WaitGroup_Unbounded", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var wg sync.WaitGroup
			for _, item := range items {
				wg.Add(1)
				go func(val int) {
					defer wg.Done()
					work(val)
				}(item)
			}
			wg.Wait()
		}
	})
}

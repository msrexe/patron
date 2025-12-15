package patron

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// ForEach executes the given function f for each element in the items slice concurrently.
// It automatically determines the optimal number of workers based on GOMAXPROCS.
// This provides a lightweight alternative to the WorkerOrchestrator for simple iteration tasks.
func ForEach[T any](items []T, f func(item T)) {
	numWorkers := runtime.GOMAXPROCS(0)
	numItems := len(items)

	// If fewer items than workers, reduce worker count to match items
	if numItems < numWorkers {
		numWorkers = numItems
	}

	if numWorkers == 0 {
		return
	}

	var idx atomic.Int64
	var wg sync.WaitGroup

	// Task closure to be executed by workers
	task := func() {
		defer wg.Done()
		for {
			// Get next index atomically
			i := int(idx.Add(1) - 1)
			if i >= numItems {
				return
			}
			f(items[i])
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go task()
	}

	wg.Wait()
}

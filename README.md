<div align="center">
		<img  src="patron.png">
    <h3>
        <strong>Patron</strong>
    </h3>
    <p>
        <strong>Patron</strong> is a high-performance Go concurrency library providing a robust worker pool and lightweight parallel iteration.
    </p>
</div>

[![Go Reference](https://pkg.go.dev/badge/github.com/msrexe/patron.svg)](https://pkg.go.dev/github.com/msrexe/patron) [![Go Report Card](https://goreportcard.com/badge/github.com/msrexe/patron)](https://goreportcard.com/report/github.com/msrexe/patron) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/msrexe/patron)

## Features

- **Worker Orchestrator**: A managed pool of workers for executing jobs with payloads and error handling. Ideal for heavy or distinct tasks.
- **ForEach Iterator**: A lightweight, high-performance parallel iterator for slices. Optimized with atomic counters and zero-allocation logic where possible.
- **Production Ready**: Proven performance and stability (see `benchmark/BENCHMARK.md`).

## Installation

```bash
go get github.com/msrexe/patron
```

## Usage

### 1. Parallel Iteration (Lightweight)

Use `patron.ForEach` to process slices concurrently with minimal overhead.

```go
data := []int{1, 2, 3, 4, 5}

// Automatically scales workers to GOMAXPROCS
patron.ForEach(data, func(n int) {
    fmt.Println(n * n)
})
```

### 2. Worker Orchestrator (Managed Pool)

Use `WorkerOrchestrator` for more complex job management including job payloads and error tracking.

```go
// Define a worker function
workerFunc := func(job *patron.Job) error {
    val, _ := job.Get("key")
    fmt.Printf("Processing %v\n", val)
    return nil
}

// Initialize orchestrator with 10 workers
orch := patron.New(patron.Config{
    WorkerCount: 10,
    WorkerFunc:  workerFunc,
})

// Queue jobs
orch.AddJobToQueue(&patron.Job{
    ID: 1, 
    Payload: map[string]interface{}{"key": "value"},
})

// Start processing and wait for results
results := orch.Start(context.Background())
```

## Performance

Patron is designed to be fast. In our benchmarks, `patron.ForEach` performs competitively with top libraries like `sourcegraph/conc` while maintaining lower memory allocations.

See [BENCHMARK.md](benchmark/BENCHMARK.md) for detailed comparisons.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](LICENSE)

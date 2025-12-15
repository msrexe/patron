# Benchmark Results

This document contains comparative performance results of the Patron library against the `conc` library, `ants`, `workerpool`, and standard Golang implementations.

## Test Environment
- **OS**: macOS (Darwin)
- **CPU**: Apple M1
- **Go Version**: 1.25.x

## 1. Iterator Comparison (`ForEach`)

Scenario: Simple arithmetic operation on a slice with 10,000 elements.

| Implementation | Ops/Sec | ns/op | B/op | allocs/op | Description |
|---|---|---|---|---|---|
| **Patron ForEach** | **~2,605** | **486,975 ns** | **120 B** | **4** | Optimized worker reuse. **Best Performance.** |
| **Conc ForEach** | ~2,404 | 487,271 ns | 328 B | 13 | Very close to Patron, slightly higher allocation. |
| **Manual WaitGroup** | ~370 | 3,544,590 ns | 487,003 B | 20,015 | Spawns 10,000 goroutines. High overhead. |
| **ErrGroup (Limited)** | ~208 | 5,922,790 ns | 480,176 B | 20,002 | Uses `SetLimit`. Still pays cost of spawning goroutines, plus locking overhead. |

**Conclusion:**
For iterating over data where each task is small, `Patron` (and `Conc`) are **~7x faster** than spawning a goroutine per item (`WaitGroup`/`ErrGroup`), thanks to worker recycling. Patron has a slight edge in memory usage.

## 2. Worker Pool Comparison

Scenario: Processing 10,000 jobs using a pool size of 100 workers.

| Implementation | ns/op | B/op | allocs/op | Note |
|---|---|---|---|---|
| **Patron Orchestrator** | **2,630,257 ns** | 2,520,212 B | 10,243 | **Fastest execution.** Higher memory usage due to Job struct wrapping. |
| **Gammazero WorkerPool**| 6,881,860 ns | 240,992 B | 10,014 | ~2.6x slower than Patron for this workload. Good memory efficiency. |
| **Ants Pool** | 8,177,407 ns | **109,184 B** | **10,197** | ~3x slower. **Best memory efficiency** (reuses gouroutines aggressively). |

**Conclusion:**
- **Patron** is optimized for **throughput** and **speed**, achieving the lowest latency for finishing the batch. It treats Jobs as objects, leading to higher memory usage (allocation per job).
- **Ants** and **WorkerPool** are optimized for **memory efficiency** and can handle millions of blocking goroutines with less footprint, but for high-throughput CPU tasks, their management overhead is higher than Patron's lightweight channel dispatch.

## 3. Patron Scalability

Patron's internal scaling under varying loads:

| Jobs | Workers | ns/op | Job Overhead |
|---|---|---|---|
| 100 | 10 | ~19 µs | ~190ns |
| 1,000 | 100 | ~262 µs | ~260ns |
| 10,000 | 500 | ~2,490 µs | ~249ns |

**Conclusion:**
Patron maintains a consistent overhead of ~250ns per job regardless of scale, proving its stability.

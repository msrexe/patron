package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/msrexe/patron"
)

func main() {
	// Generate some data
	data := make([]int, 20)
	for i := range data {
		data[i] = rand.Intn(100)
	}

	fmt.Println("Processing data in parallel using Patron ForEach...")

	// Process data concurrently
	patron.ForEach(data, func(n int) {
		// Simulate some CPU intensive work
		result := expensiveCalculation(n)
		fmt.Printf("Processed input: %d -> result: %d\n", n, result)
	})

	fmt.Println("All data processed!")
}

func expensiveCalculation(n int) int {
	time.Sleep(100 * time.Millisecond) // Simulate delay
	return n * n
}

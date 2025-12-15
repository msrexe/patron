package patron

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestForEach(t *testing.T) {
	t.Run("basic execution", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		var sum int64

		ForEach(items, func(item int) {
			atomic.AddInt64(&sum, int64(item))
		})

		if sum != 15 {
			t.Errorf("Expected sum 15, got %d", sum)
		}
	})

	t.Run("concurrent constraints", func(t *testing.T) {
		// This test ensures that we are actually running concurrently
		// If it was sequential, it would take 500ms
		// With concurrency, it should be much faster (approx 100ms)
		items := []int{1, 2, 3, 4, 5}
		start := time.Now()

		ForEach(items, func(item int) {
			time.Sleep(100 * time.Millisecond)
		})

		duration := time.Since(start)
		// Should be less than sum of sleeps (e.g. < 400ms)
		if duration >= 400*time.Millisecond {
			t.Errorf("Expected duration < 400ms, got %v", duration)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var items []int
		ForEach(items, func(item int) {})
	})
}

func BenchmarkForEach(b *testing.B) {
	// Setup a large slice
	count := 10000
	items := make([]int, count)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	b.Run("PatronForEach", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ForEach(items, func(item int) {
				_ = item * 2
			})
		}
	})
}

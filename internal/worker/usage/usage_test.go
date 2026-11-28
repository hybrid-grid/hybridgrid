package usage

import (
	"context"
	"testing"
)

func TestSampleReadsThisMachine(t *testing.T) {
	var s Sampler
	for i := 0; i < 2; i++ { // first call primes the CPU window, second is a delta
		got, err := s.Sample(context.Background())
		if err != nil {
			t.Fatalf("Sample #%d: %v", i, err)
		}
		if got.MemoryTotalBytes == 0 {
			t.Fatalf("Sample #%d: zero total memory", i)
		}
		if got.MemoryUsedBytes > got.MemoryTotalBytes {
			t.Fatalf("Sample #%d: used %d > total %d", i, got.MemoryUsedBytes, got.MemoryTotalBytes)
		}
		if got.CPUPercent < 0 || got.CPUPercent > 100 {
			t.Fatalf("Sample #%d: cpu %v out of range", i, got.CPUPercent)
		}
		if got.At.IsZero() {
			t.Fatalf("Sample #%d: zero timestamp", i)
		}
	}
}

func TestClampPercent(t *testing.T) {
	for in, want := range map[float64]float64{-3: 0, 0: 0, 42.5: 42.5, 100: 100, 130: 100} {
		if got := clampPercent(in); got != want {
			t.Errorf("clampPercent(%v) = %v, want %v", in, got, want)
		}
	}
}

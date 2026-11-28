// Package usage samples the worker machine's live CPU and memory use, for
// the dashboard's per-worker resource charts.
package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// Sample is one reading of the whole machine (not just this process).
type Sample struct {
	At time.Time
	// CPUPercent is system-wide utilisation across all cores, 0..100.
	CPUPercent float64
	// MemoryUsedBytes is Total - Available: memory the OS cannot hand to a
	// new process right now. gopsutil's Used would also count reclaimable
	// file cache, which on macOS makes every machine look nearly full.
	MemoryUsedBytes  uint64
	MemoryTotalBytes uint64
	// Load1 is the 1-minute load average; 0 where unsupported (Windows).
	Load1 float64
}

// Sampler takes readings. CPU percent is measured over the interval since
// the previous Sample call, so the first reading of a fresh Sampler is
// taken over a short blocking window instead.
type Sampler struct {
	primed bool
}

// firstWindow is how long the very first CPU reading blocks to measure.
const firstWindow = 200 * time.Millisecond

// Sample reads CPU, memory and load. Memory or CPU failures are errors;
// a missing load average is not.
func (s *Sampler) Sample(ctx context.Context) (Sample, error) {
	window := time.Duration(0)
	if !s.primed {
		window = firstWindow
	}
	percents, err := cpu.PercentWithContext(ctx, window, false)
	if err != nil {
		return Sample{}, fmt.Errorf("failed to read cpu usage: %w", err)
	}
	s.primed = true
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return Sample{}, fmt.Errorf("failed to read memory usage: %w", err)
	}
	out := Sample{At: time.Now(), MemoryTotalBytes: vm.Total}
	if len(percents) > 0 {
		out.CPUPercent = clampPercent(percents[0])
	}
	if vm.Total >= vm.Available {
		out.MemoryUsedBytes = vm.Total - vm.Available
	}
	if avg, err := load.AvgWithContext(ctx); err == nil {
		out.Load1 = avg.Load1
	}
	return out, nil
}

func clampPercent(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 100:
		return 100
	default:
		return v
	}
}

package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	"github.com/h3nr1-d14z/hybridgrid/internal/worker/usage"
)

// defaultUsageInterval is how often a worker reports CPU/memory use: often
// enough for a live chart, cheap enough to ignore (one small unary RPC).
const defaultUsageInterval = 2 * time.Second

// runUsageReporter samples this machine's CPU and memory every interval
// and reports it to the coordinator until stop closes. Failures are
// logged at debug level only: the heartbeat loop already warns when the
// coordinator is unreachable, and this would repeat it every 2 seconds.
func runUsageReporter(cli *client.Client, workerID, token string, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		return
	}
	var sampler usage.Sampler
	report := func() {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		defer cancel()
		s, err := sampler.Sample(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Usage sample failed")
			return
		}
		if _, err := cli.ReportUsage(ctx, &pb.WorkerUsageReport{
			WorkerId:         workerID,
			AuthToken:        token,
			CpuPercent:       s.CPUPercent,
			MemoryUsedBytes:  s.MemoryUsedBytes,
			MemoryTotalBytes: s.MemoryTotalBytes,
			Load1:            s.Load1,
			SampledAtMs:      s.At.UnixMilli(),
		}); err != nil {
			log.Debug().Err(err).Msg("Usage report failed")
		}
	}

	report()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			report()
		case <-stop:
			return
		}
	}
}

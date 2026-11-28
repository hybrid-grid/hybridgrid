package server

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

func TestReportUsage(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, AuthToken: "secret"})
	defer cleanup()
	tele := telemetry.NewService(telemetry.NewStore(), s.NewStatsProvider())
	s.SetTelemetry(tele)
	require.NoError(t, s.registry.Add(&registry.WorkerInfo{
		ID:           "w1",
		Address:      "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{Hostname: "box", MaxParallelTasks: 2},
		MaxParallel:  2,
	}))

	_, err := client.ReportUsage(context.Background(), &pb.WorkerUsageReport{WorkerId: "w1", AuthToken: "nope"})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	resp, err := client.ReportUsage(context.Background(), &pb.WorkerUsageReport{WorkerId: "ghost", AuthToken: "secret"})
	require.NoError(t, err)
	assert.False(t, resp.Accepted, "unregistered workers are not recorded")

	resp, err = client.ReportUsage(context.Background(), &pb.WorkerUsageReport{
		WorkerId: "w1", AuthToken: "secret", CpuPercent: 37.5, MemoryUsedBytes: 6 << 30, MemoryTotalBytes: 16 << 30, Load1: 2.25,
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	got, ok := tele.LatestUsage("w1")
	require.True(t, ok)
	assert.Equal(t, 37.5, got.CPUPercent)
	assert.Equal(t, uint64(6<<30), got.MemoryUsedBytes)
	assert.WithinDuration(t, time.Now(), time.UnixMilli(got.AtMs), 5*time.Second, "stamped with the coordinator clock")

	// The latest reading rides along on the worker listing.
	workers := s.NewStatsProvider().GetWorkers()
	require.Len(t, workers, 1)
	assert.Equal(t, 37.5, workers[0].CPUPercent)
	assert.Equal(t, uint64(16<<30), workers[0].MemoryTotalBytes)
	assert.Equal(t, got.AtMs, workers[0].UsageAtMs)
}

func TestSanitizeUsage(t *testing.T) {
	now := time.UnixMilli(1_000)
	tests := []struct {
		name string
		in   *pb.WorkerUsageReport
		want telemetry.UsageSample
	}{
		{"passes valid values", &pb.WorkerUsageReport{CpuPercent: 50, MemoryUsedBytes: 1, MemoryTotalBytes: 2, Load1: 1.5},
			telemetry.UsageSample{AtMs: 1_000, CPUPercent: 50, MemoryUsedBytes: 1, MemoryTotalBytes: 2, Load1: 1.5}},
		{"clamps cpu above 100", &pb.WorkerUsageReport{CpuPercent: 250}, telemetry.UsageSample{AtMs: 1_000, CPUPercent: 100}},
		{"drops NaN and negatives", &pb.WorkerUsageReport{CpuPercent: math.NaN(), Load1: -1}, telemetry.UsageSample{AtMs: 1_000}},
		{"used never exceeds total", &pb.WorkerUsageReport{MemoryUsedBytes: 9, MemoryTotalBytes: 4}, telemetry.UsageSample{AtMs: 1_000, MemoryUsedBytes: 4, MemoryTotalBytes: 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeUsage(tt.in, now))
		})
	}
}

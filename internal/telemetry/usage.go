package telemetry

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// UsageHistory is how many samples are kept per worker: five minutes at
// the workers' default 2-second reporting interval.
const UsageHistory = 150

// UsageSample is one live CPU/memory reading of a worker machine.
type UsageSample struct {
	AtMs             int64   `json:"at_ms"`
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryUsedBytes  uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes uint64  `json:"memory_total_bytes"`
	Load1            float64 `json:"load1"`
}

// RecordUsage keeps a worker's reading in its bounded history and pushes
// it to live subscribers as a "worker_usage" event. Usage events are not
// added to the replay ring: ListWorkerUsage serves the backlog.
func (s *Service) RecordUsage(workerID string, sample UsageSample) {
	s.usageMu.Lock()
	if s.usage == nil {
		s.usage = make(map[string][]UsageSample)
	}
	history := append(s.usage[workerID], sample)
	if len(history) > UsageHistory {
		history = history[len(history)-UsageHistory:]
	}
	s.usage[workerID] = history
	s.usageMu.Unlock()

	payload, err := json.Marshal(struct {
		WorkerID string `json:"worker_id"`
		UsageSample
	}{workerID, sample})
	if err != nil {
		log.Error().Err(err).Msg("telemetry: failed to marshal usage event")
		return
	}
	s.broadcast(&pb.TelemetryEvent{Type: "worker_usage", Timestamp: time.Now().Unix(), PayloadJson: string(payload)})
}

// LatestUsage returns a worker's most recent reading, if any.
func (s *Service) LatestUsage(workerID string) (UsageSample, bool) {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	history := s.usage[workerID]
	if len(history) == 0 {
		return UsageSample{}, false
	}
	return history[len(history)-1], true
}

// ListWorkerUsage returns every reporting worker's recent history.
func (s *Service) ListWorkerUsage(_ context.Context, req *pb.ListTelemetryWorkerUsageRequest) (*pb.ListTelemetryWorkerUsageResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	ids := make([]string, 0, len(s.usage))
	for id := range s.usage {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]*pb.TelemetryWorkerUsage, 0, len(ids))
	for _, id := range ids {
		samples := make([]*pb.TelemetryUsageSample, 0, len(s.usage[id]))
		for _, u := range s.usage[id] {
			samples = append(samples, &pb.TelemetryUsageSample{
				AtMs:             u.AtMs,
				CpuPercent:       u.CPUPercent,
				MemoryUsedBytes:  u.MemoryUsedBytes,
				MemoryTotalBytes: u.MemoryTotalBytes,
				Load1:            u.Load1,
			})
		}
		out = append(out, &pb.TelemetryWorkerUsage{WorkerId: id, Samples: samples})
	}
	return &pb.ListTelemetryWorkerUsageResponse{Workers: out}, nil
}

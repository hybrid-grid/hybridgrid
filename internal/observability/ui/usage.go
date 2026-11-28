package ui

import (
	"encoding/json"
	"net/http"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// UsageSample is one live CPU/memory reading as the SPA sees it.
type UsageSample struct {
	AtMs             int64   `json:"at_ms"`
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryUsedBytes  uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes uint64  `json:"memory_total_bytes"`
	Load1            float64 `json:"load1"`
}

// WorkerUsage is one worker's recent history, oldest sample first.
type WorkerUsage struct {
	WorkerID string        `json:"worker_id"`
	Samples  []UsageSample `json:"samples"`
}

// handleWorkerUsage serves every reporting worker's recent CPU/memory
// history, the backlog charts start from before live events arrive.
func (s *Server) handleWorkerUsage(w http.ResponseWriter, r *http.Request) {
	resp, err := s.client.ListWorkerUsage(r.Context(), &pb.ListTelemetryWorkerUsageRequest{AuthToken: s.config.AuthToken})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	workers := make([]WorkerUsage, 0, len(resp.GetWorkers()))
	for _, wu := range resp.GetWorkers() {
		samples := make([]UsageSample, 0, len(wu.GetSamples()))
		for _, u := range wu.GetSamples() {
			samples = append(samples, UsageSample{
				AtMs:             u.GetAtMs(),
				CPUPercent:       u.GetCpuPercent(),
				MemoryUsedBytes:  u.GetMemoryUsedBytes(),
				MemoryTotalBytes: u.GetMemoryTotalBytes(),
				Load1:            u.GetLoad1(),
			})
		}
		workers = append(workers, WorkerUsage{WorkerID: wu.GetWorkerId(), Samples: samples})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"workers": workers})
}

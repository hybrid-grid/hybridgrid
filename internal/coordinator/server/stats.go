package server

import (
	"sync/atomic"
	"time"

	"github.com/h3nr1-d14z/hybridgrid/internal/observability/dashboard"
)

// statsProvider implements dashboard.StatsProvider for the coordinator.
type statsProvider struct {
	server    *Server
	startTime time.Time
}

// NewStatsProvider creates a new stats provider for the coordinator.
func (s *Server) NewStatsProvider() dashboard.StatsProvider {
	return &statsProvider{
		server:    s,
		startTime: time.Now(),
	}
}

// GetStats returns current cluster statistics.
func (p *statsProvider) GetStats() *dashboard.Stats {
	workers := p.server.registry.List()

	healthyCount := 0
	for _, w := range workers {
		if w.IsHealthy(p.server.config.HeartbeatTTL) {
			healthyCount++
		}
	}

	// Calculate cache hit rate
	cacheHits := atomic.LoadInt64(&p.server.cacheHits)
	cacheMisses := atomic.LoadInt64(&p.server.cacheMisses)
	cacheTotal := cacheHits + cacheMisses
	cacheHitRate := 0.0
	if cacheTotal > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheTotal)
	}

	return &dashboard.Stats{
		TotalTasks:     atomic.LoadInt64(&p.server.totalTasks),
		SuccessTasks:   atomic.LoadInt64(&p.server.successTasks),
		FailedTasks:    atomic.LoadInt64(&p.server.failedTasks),
		ActiveTasks:    atomic.LoadInt64(&p.server.activeTasks),
		QueuedTasks:    atomic.LoadInt64(&p.server.queuedTasks),
		CacheHits:      cacheHits,
		CacheMisses:    cacheMisses,
		CacheHitRate:   cacheHitRate,
		TotalWorkers:   len(workers),
		HealthyWorkers: healthyCount,
		UptimeSeconds:  int64(time.Since(p.startTime).Seconds()),
		Timestamp:      time.Now().Unix(),
	}
}

// GetWorkers returns current worker information.
func (p *statsProvider) GetWorkers() []*dashboard.WorkerInfo {
	workers := p.server.registry.List()
	result := make([]*dashboard.WorkerInfo, 0, len(workers))

	for _, w := range workers {
		// Calculate success rate
		successRate := 0.0
		if w.TotalTasks > 0 {
			successRate = float64(w.SuccessfulTasks) / float64(w.TotalTasks)
		}

		// Get circuit state
		circuitState := "CLOSED"
		if p.server.circuitManager != nil {
			state := p.server.circuitManager.GetState(w.ID)
			circuitState = string(state)
		}

		info := &dashboard.WorkerInfo{
			ID:              w.ID,
			Host:            w.Capabilities.Hostname,
			Address:         w.Address,
			Architecture:    w.Capabilities.NativeArch.String(),
			CPUCores:        w.Capabilities.CpuCores,
			MemoryGB:        float64(w.Capabilities.MemoryBytes) / (1024 * 1024 * 1024),
			ActiveTasks:     w.ActiveTasks,
			TotalTasks:      w.TotalTasks,
			SuccessRate:     successRate,
			AvgLatencyMs:    float64(w.AvgCompileTime.Milliseconds()),
			CircuitState:    circuitState,
			DiscoverySource: w.DiscoverySource,
			Healthy:         w.IsHealthy(p.server.config.HeartbeatTTL),
			LastSeen:        w.LastHeartbeat.Unix(),
		}
		result = append(result, info)
	}

	return result
}

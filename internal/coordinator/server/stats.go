package server

import (
	"sync/atomic"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
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

	cacheHits := atomic.LoadInt64(&p.server.cacheHits)
	cacheMisses := atomic.LoadInt64(&p.server.cacheMisses)
	cacheTotal := cacheHits + cacheMisses
	cacheHitRate := 0.0
	if cacheTotal > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheTotal)
	}

	flutterCacheHits := atomic.LoadInt64(&p.server.flutterCacheHits)
	flutterCacheMisses := atomic.LoadInt64(&p.server.flutterCacheMisses)
	flutterCacheTotal := flutterCacheHits + flutterCacheMisses
	flutterCacheHitRate := 0.0
	if flutterCacheTotal > 0 {
		flutterCacheHitRate = float64(flutterCacheHits) / float64(flutterCacheTotal)
	}

	unityCacheHits := atomic.LoadInt64(&p.server.unityCacheHits)
	unityCacheMisses := atomic.LoadInt64(&p.server.unityCacheMisses)
	unityCacheTotal := unityCacheHits + unityCacheMisses
	unityCacheHitRate := 0.0
	if unityCacheTotal > 0 {
		unityCacheHitRate = float64(unityCacheHits) / float64(unityCacheTotal)
	}

	return &dashboard.Stats{
		TotalTasks:          atomic.LoadInt64(&p.server.totalTasks),
		SuccessTasks:        atomic.LoadInt64(&p.server.successTasks),
		FailedTasks:         atomic.LoadInt64(&p.server.failedTasks),
		ActiveTasks:         atomic.LoadInt64(&p.server.activeTasks),
		QueuedTasks:         atomic.LoadInt64(&p.server.queuedTasks),
		CacheHits:           cacheHits,
		CacheMisses:         cacheMisses,
		CacheHitRate:        cacheHitRate,
		FlutterBuilds:       atomic.LoadInt64(&p.server.flutterBuilds),
		FlutterCacheHits:    flutterCacheHits,
		FlutterCacheMisses:  flutterCacheMisses,
		FlutterCacheHitRate: flutterCacheHitRate,
		UnityBuilds:         atomic.LoadInt64(&p.server.unityBuilds),
		UnityCacheHits:      unityCacheHits,
		UnityCacheMisses:    unityCacheMisses,
		UnityCacheHitRate:   unityCacheHitRate,
		TotalWorkers:        len(workers),
		HealthyWorkers:      healthyCount,
		UptimeSeconds:       int64(time.Since(p.startTime).Seconds()),
		Timestamp:           time.Now().Unix(),
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

		caps := workerCapabilitiesOrDefault(w)
		cppCaps := caps.GetCpp()
		compilers := make([]string, 0)
		if cppCaps != nil {
			compilers = append(compilers, cppCaps.GetCompilers()...)
		}

		info := &dashboard.WorkerInfo{
			ID:                w.ID,
			Host:              caps.Hostname,
			Address:           w.Address,
			OS:                caps.Os,
			Architecture:      caps.NativeArch.String(),
			Architectures:     supportedArchitectures(caps),
			CPUCores:          caps.CpuCores,
			MemoryGB:          float64(caps.MemoryBytes) / (1024 * 1024 * 1024),
			MaxParallelTasks:  caps.MaxParallelTasks,
			ActiveTasks:       w.ActiveTasks,
			TotalTasks:        w.TotalTasks,
			SuccessRate:       successRate,
			AvgLatencyMs:      float64(w.AvgCompileTime.Milliseconds()),
			CircuitState:      circuitState,
			DiscoverySource:   w.DiscoverySource,
			Version:           caps.Version,
			DockerAvailable:   caps.DockerAvailable,
			FlutterAvailable:  caps.GetFlutter() != nil,
			FlutterSDKVersion: flutterSDKVersion(caps),
			FlutterPlatforms:  flutterPlatforms(caps),
			UnityAvailable:    caps.GetUnity() != nil,
			UnityVersions:     unityVersions(caps),
			UnityPlatforms:    unityPlatforms(caps),
			Compilers:         compilers,
			BuildTypes:        supportedBuildTypes(caps),
			Healthy:           w.IsHealthy(p.server.config.HeartbeatTTL),
			LastSeen:          w.LastHeartbeat.Unix(),
		}
		result = append(result, info)
	}

	return result
}

func supportedArchitectures(caps *pb.WorkerCapabilities) []string {
	architectures := make([]string, 0, 1)
	seen := make(map[string]struct{})

	add := func(value string) {
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		architectures = append(architectures, value)
	}

	add(caps.NativeArch.String())

	if cppCaps := caps.GetCpp(); cppCaps != nil {
		for _, arch := range cppCaps.GetMsvcArchitectures() {
			add(arch)
		}
	}

	return architectures
}

func supportedBuildTypes(caps *pb.WorkerCapabilities) []string {
	buildTypes := make([]string, 0, 6)

	if caps.GetCpp() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_CPP.String())
	}
	if caps.GetFlutter() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_FLUTTER.String())
	}
	if caps.GetUnity() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_UNITY.String())
	}
	if caps.GetRust() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_RUST.String())
	}
	if caps.GetGo() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_GO.String())
	}
	if caps.GetNodejs() != nil {
		buildTypes = append(buildTypes, pb.BuildType_BUILD_TYPE_NODEJS.String())
	}

	return buildTypes
}

func flutterSDKVersion(caps *pb.WorkerCapabilities) string {
	if caps.GetFlutter() != nil {
		return caps.GetFlutter().GetSdkVersion()
	}
	return ""
}

func flutterPlatforms(caps *pb.WorkerCapabilities) []string {
	if caps.GetFlutter() == nil {
		return nil
	}
	platforms := caps.GetFlutter().GetPlatforms()
	result := make([]string, len(platforms))
	for i, p := range platforms {
		result[i] = p.String()
	}
	return result
}

func unityVersions(caps *pb.WorkerCapabilities) []string {
	if caps.GetUnity() == nil {
		return nil
	}
	return caps.GetUnity().GetVersions()
}

func unityPlatforms(caps *pb.WorkerCapabilities) []string {
	if caps.GetUnity() == nil {
		return nil
	}
	targets := caps.GetUnity().GetBuildTargets()
	result := make([]string, len(targets))
	for i, t := range targets {
		result[i] = t.String()
	}
	return result
}

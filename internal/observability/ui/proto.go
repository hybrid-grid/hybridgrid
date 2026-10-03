package ui

import (
	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// The gRPC transport reshapes the telemetry wire types into generated
// proto messages. The REST/WS surface must keep emitting the original
// JSON tag layouts (internal/telemetry), so these converters rebuild
// those structs field-by-field on every ingress. They are the only
// place a field can be dropped silently during a proto regen, so keep
// them exhaustive against the protobuf message set.

func statsFromProto(s *pb.TelemetryStats) *Stats {
	if s == nil {
		return &Stats{}
	}
	return &Stats{
		TotalTasks:          s.GetTotalTasks(),
		SuccessTasks:        s.GetSuccessTasks(),
		FailedTasks:         s.GetFailedTasks(),
		ActiveTasks:         s.GetActiveTasks(),
		QueuedTasks:         s.GetQueuedTasks(),
		CacheHits:           s.GetCacheHits(),
		CacheMisses:         s.GetCacheMisses(),
		CacheHitRate:        s.GetCacheHitRate(),
		FlutterBuilds:       s.GetFlutterBuilds(),
		FlutterCacheHits:    s.GetFlutterCacheHits(),
		FlutterCacheMisses:  s.GetFlutterCacheMisses(),
		FlutterCacheHitRate: s.GetFlutterCacheHitRate(),
		UnityBuilds:         s.GetUnityBuilds(),
		UnityCacheHits:      s.GetUnityCacheHits(),
		UnityCacheMisses:    s.GetUnityCacheMisses(),
		UnityCacheHitRate:   s.GetUnityCacheHitRate(),
		TotalWorkers:        int(s.GetTotalWorkers()),
		HealthyWorkers:      int(s.GetHealthyWorkers()),
		UptimeSeconds:       s.GetUptimeSeconds(),
		Timestamp:           s.GetTimestamp(),
	}
}

func workerFromProto(w *pb.TelemetryWorker) *WorkerInfo {
	if w == nil {
		return &WorkerInfo{}
	}
	return &WorkerInfo{
		ID:                w.GetId(),
		Host:              w.GetHost(),
		Address:           w.GetAddress(),
		OS:                w.GetOs(),
		Architecture:      w.GetArchitecture(),
		Architectures:     w.GetArchitectures(),
		CPUCores:          w.GetCpuCores(),
		MemoryGB:          w.GetMemoryGb(),
		MaxParallelTasks:  w.GetMaxParallelTasks(),
		ActiveTasks:       w.GetActiveTasks(),
		TotalTasks:        w.GetTotalTasks(),
		SuccessRate:       w.GetSuccessRate(),
		AvgLatencyMs:      w.GetAvgLatencyMs(),
		CircuitState:      w.GetCircuitState(),
		DiscoverySource:   w.GetDiscoverySource(),
		Version:           w.GetVersion(),
		DockerAvailable:   w.GetDockerAvailable(),
		FlutterAvailable:  w.GetFlutterAvailable(),
		FlutterSDKVersion: w.GetFlutterSdkVersion(),
		FlutterPlatforms:  w.GetFlutterPlatforms(),
		UnityAvailable:    w.GetUnityAvailable(),
		UnityVersions:     w.GetUnityVersions(),
		UnityPlatforms:    w.GetUnityPlatforms(),
		Compilers:         w.GetCompilers(),
		BuildTypes:        w.GetBuildTypes(),
		Healthy:           w.GetHealthy(),
		LastSeen:          w.GetLastSeen(),
	}
}

func taskFromProto(t *pb.TelemetryTask) *TaskInfo {
	if t == nil {
		return &TaskInfo{}
	}
	return &TaskInfo{
		ID:            t.GetId(),
		BuildType:     t.GetBuildType(),
		BuildID:       t.GetBuildId(),
		Status:        t.GetStatus(),
		WorkerID:      t.GetWorkerId(),
		StartedAtMs:   t.GetStartedAtMs(),
		CompletedAtMs: t.GetCompletedAtMs(),
		DurationMs:    t.GetDurationMs(),
		QueueMs:       t.GetQueueMs(),
		CompileMs:     t.GetCompileMs(),
		ExitCode:      t.GetExitCode(),
		FromCache:     t.GetFromCache(),
		ErrorMessage:  t.GetErrorMessage(),
	}
}

func buildFromProto(b *pb.TelemetryBuild) *BuildInfo {
	if b == nil {
		return &BuildInfo{}
	}
	return &BuildInfo{
		ID:             b.GetId(),
		BuildType:      b.GetBuildType(),
		Status:         b.GetStatus(),
		TotalTasks:     int(b.GetTotalTasks()),
		CompletedTasks: int(b.GetCompletedTasks()),
		FailedTasks:    int(b.GetFailedTasks()),
		RunningTasks:   int(b.GetRunningTasks()),
		FromCacheCount: int(b.GetFromCacheCount()),
		FirstTaskAtMs:  b.GetFirstTaskAtMs(),
		LastTaskAtMs:   b.GetLastTaskAtMs(),
		Truncated:      b.GetTruncated(),
	}
}

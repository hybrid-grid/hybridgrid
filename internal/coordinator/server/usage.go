package server

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

// ReportUsage records a worker's live CPU/memory reading for the
// dashboard. It is authenticated like Handshake (it comes from workers),
// accepts only registered workers, and stamps samples with the
// coordinator's clock so charts of different machines share one time
// axis regardless of worker clock skew.
func (s *Server) ReportUsage(_ context.Context, req *pb.WorkerUsageReport) (*pb.ReportUsageResponse, error) {
	if s.config.AuthToken != "" && req.GetAuthToken() != s.config.AuthToken {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}
	if req.GetWorkerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "worker_id required")
	}
	if _, ok := s.registry.Get(req.GetWorkerId()); !ok {
		return &pb.ReportUsageResponse{Accepted: false}, nil
	}
	if s.telemetry != nil {
		s.telemetry.RecordUsage(req.GetWorkerId(), sanitizeUsage(req, time.Now()))
	}
	return &pb.ReportUsageResponse{Accepted: true}, nil
}

func sanitizeUsage(req *pb.WorkerUsageReport, now time.Time) telemetry.UsageSample {
	cpu := req.GetCpuPercent()
	if math.IsNaN(cpu) || cpu < 0 {
		cpu = 0
	} else if cpu > 100 {
		cpu = 100
	}
	load1 := req.GetLoad1()
	if math.IsNaN(load1) || load1 < 0 {
		load1 = 0
	}
	used, total := req.GetMemoryUsedBytes(), req.GetMemoryTotalBytes()
	if used > total {
		used = total
	}
	return telemetry.UsageSample{
		AtMs:             now.UnixMilli(),
		CPUPercent:       cpu,
		MemoryUsedBytes:  used,
		MemoryTotalBytes: total,
		Load1:            load1,
	}
}

// leastloaded-plugin is the reference external scheduler plugin: it
// implements the schedulerv1alpha1 gRPC contract
// (proto/hybridgrid/v1/scheduler_plugin.proto) with a least-loaded
// policy over the fleet snapshot the coordinator sends with each
// Select. Run it, then start the coordinator with:
//
//	hg-coord serve --scheduler=external \
//	  --scheduler-plugin-addr=localhost:50099
//
// It demonstrates everything a real plugin needs: reading the worker
// snapshots, answering within the coordinator's timeout, and consuming
// Feedback events.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

type server struct {
	pb.UnimplementedSchedulerPluginServer
}

// Select picks the worker with the most spare capacity among the
// healthy, circuit-closed workers that serve the requested build type.
func (s *server) Select(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
	var best *pb.WorkerSnapshot
	bestRatio := 2.0
	for _, w := range req.GetWorkers() {
		if !w.GetHealthy() || w.GetCircuitOpen() {
			continue
		}
		if !supports(w.GetBuildTypes(), req.GetBuildType()) {
			continue
		}
		maxParallel := float64(w.GetMaxParallel())
		if maxParallel <= 0 {
			maxParallel = 1
		}
		ratio := float64(w.GetActiveTasks()) / maxParallel
		if ratio < bestRatio {
			bestRatio = ratio
			best = w
		}
	}
	if best == nil {
		// Empty answer: the coordinator falls back to its inner
		// scheduler (leastloaded) — a plugin may decline to decide.
		return &pb.SelectResponse{}, nil
	}
	log.Printf("select: build=%s workers=%d -> %s (load %.2f)",
		req.GetBuildType(), len(req.GetWorkers()), best.GetId(), bestRatio)
	return &pb.SelectResponse{WorkerId: best.GetId()}, nil
}

// Feedback logs completions; a learning plugin would update its model.
func (s *server) Feedback(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error) {
	log.Printf("feedback: worker=%s success=%v duration=%.1fms",
		req.GetWorkerId(), req.GetSuccess(), req.GetDurationMs())
	return &pb.FeedbackResponse{}, nil
}

func supports(buildTypes []string, want string) bool {
	for _, t := range buildTypes {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

func main() {
	port := flag.Int("port", 50099, "listen port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterSchedulerPluginServer(gs, &server{})
	log.Printf("leastloaded scheduler plugin listening on %s", lis.Addr())
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

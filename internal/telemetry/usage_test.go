package telemetry

import (
	"context"
	"encoding/json"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestRecordUsageKeepsBoundedHistory(t *testing.T) {
	svc := NewService(NewStore(), nil)
	for i := 0; i < UsageHistory+10; i++ {
		svc.RecordUsage("w2", UsageSample{AtMs: int64(i), CPUPercent: float64(i % 100)})
	}
	svc.RecordUsage("w1", UsageSample{AtMs: 7})

	latest, ok := svc.LatestUsage("w2")
	if !ok || latest.AtMs != UsageHistory+9 {
		t.Fatalf("LatestUsage = %+v, %v", latest, ok)
	}
	if _, ok := svc.LatestUsage("none"); ok {
		t.Fatal("LatestUsage for an unknown worker should report false")
	}

	resp, err := svc.ListWorkerUsage(context.Background(), &pb.ListTelemetryWorkerUsageRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Workers) != 2 || resp.Workers[0].WorkerId != "w1" || resp.Workers[1].WorkerId != "w2" {
		t.Fatalf("workers not listed sorted by id: %+v", resp.Workers)
	}
	samples := resp.Workers[1].Samples
	if len(samples) != UsageHistory || samples[0].AtMs != 10 || samples[len(samples)-1].AtMs != UsageHistory+9 {
		t.Fatalf("history = %d samples, first %d, last %d", len(samples), samples[0].AtMs, samples[len(samples)-1].AtMs)
	}
}

func TestRecordUsageBroadcastsLiveEvent(t *testing.T) {
	svc := NewService(NewStore(), nil)
	ch := make(chan *pb.TelemetryEvent, 1)
	svc.subsMu.Lock()
	svc.subs[ch] = struct{}{}
	svc.subsMu.Unlock()

	svc.RecordUsage("w1", UsageSample{AtMs: 5, CPUPercent: 12.5, MemoryUsedBytes: 3, MemoryTotalBytes: 8})
	ev := <-ch
	if ev.Type != "worker_usage" {
		t.Fatalf("event type = %q", ev.Type)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(ev.PayloadJson), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["worker_id"] != "w1" || payload["cpu_percent"] != 12.5 || payload["at_ms"] != float64(5) {
		t.Fatalf("payload = %v", payload)
	}
}

func TestListWorkerUsageRequiresToken(t *testing.T) {
	svc := NewService(NewStore(), nil)
	svc.Token = "secret"
	if _, err := svc.ListWorkerUsage(context.Background(), &pb.ListTelemetryWorkerUsageRequest{AuthToken: "nope"}); err == nil {
		t.Fatal("expected an auth error")
	}
}

package flutter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/taskid"
)

// defaultABIs is Flutter's default Android release set, one task each.
var defaultABIs = []string{"android-arm64", "android-arm", "android-x64"}

// normalizeABIs accepts "arm64", "android-arm64", "x64", … and returns the
// canonical --target-platform values in order, without duplicates.
func normalizeABIs(values []string) ([]string, error) {
	if len(values) == 0 {
		return append([]string(nil), defaultABIs...), nil
	}
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, v := range values {
		abi := strings.ToLower(strings.TrimSpace(v))
		if abi == "" {
			continue
		}
		if !strings.HasPrefix(abi, "android-") {
			abi = "android-" + abi
		}
		switch abi {
		case "android-arm", "android-arm64", "android-x64":
		default:
			return nil, fmt.Errorf("unsupported ABI %q (supported: arm64, arm, x64)", v)
		}
		if !seen[abi] {
			seen[abi] = true
			out = append(out, abi)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no ABI selected")
	}
	return out, nil
}

// abiResult is one per-ABI task of a distributed build.
type abiResult struct {
	ABI  string
	Resp *pb.BuildResponse
	Err  error
	Took time.Duration
}

// abiRequests splits one APK build into one task per ABI. All tasks share
// the source archive and the build id, so the coordinator schedules them
// as one build — on different workers when the fleet has room — and the
// dashboard shows them together.
func abiRequests(base *pb.BuildRequest, abis []string) []*pb.BuildRequest {
	reqs := make([]*pb.BuildRequest, 0, len(abis))
	for _, abi := range abis {
		req := proto.Clone(base).(*pb.BuildRequest)
		req.TaskId = taskid.NewTaskID()
		cfg := req.GetFlutterConfig()
		cfg.TargetPlatforms = []string{abi}
		cfg.SplitPerAbi = true
		reqs = append(reqs, req)
	}
	return reqs
}

// runDistributed sends every per-ABI task at once over one connection and
// waits for all of them; results come back in ABI order.
func runDistributed(ctx context.Context, client BuildClient, reqs []*pb.BuildRequest) []abiResult {
	results := make([]abiResult, len(reqs))
	var wg sync.WaitGroup
	for i, req := range reqs {
		wg.Add(1)
		go func(i int, req *pb.BuildRequest) {
			defer wg.Done()
			start := time.Now()
			resp, err := client.Build(ctx, req)
			results[i] = abiResult{ABI: req.GetFlutterConfig().GetTargetPlatforms()[0], Resp: resp, Err: err, Took: time.Since(start)}
		}(i, req)
	}
	wg.Wait()
	sort.SliceStable(results, func(a, b int) bool { return abiIndex(results[a].ABI) < abiIndex(results[b].ABI) })
	return results
}

func abiIndex(abi string) int {
	for i, d := range defaultABIs {
		if d == abi {
			return i
		}
	}
	return len(defaultABIs)
}

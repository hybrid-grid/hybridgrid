package capability

import (
	"errors"
	"testing"
)

// fakeFS builds a fileReader from a map of path -> content, returning
// os.ErrNotExist for any path not in the map — matching os.ReadFile's
// error behavior closely enough for the "file doesn't exist, try the
// next cgroup version" fallback logic under test.
func fakeFS(files map[string]string) fileReader {
	return func(path string) ([]byte, error) {
		if content, ok := files[path]; ok {
			return []byte(content), nil
		}
		return nil, errors.New("file not found: " + path)
	}
}

func TestDetectCPUMillisFrom_CgroupV2_HalfCore(t *testing.T) {
	// Docker `--cpus=0.5` under cgroup v2: quota 50000us / period 100000us.
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/cpu.max": "50000 100000\n",
	})
	got := detectCPUMillisFrom(read)
	if got != 500 {
		t.Errorf("got %d, want 500 (0.5 cores)", got)
	}
}

func TestDetectCPUMillisFrom_CgroupV2_FractionalCores(t *testing.T) {
	// Mirrors the actual test/stress/docker-compose-hetero.yml quotas.
	cases := []struct {
		quotaPeriod string
		wantMillis  int32
	}{
		{"50000 100000", 500},  // --cpus=0.5
		{"60000 100000", 600},  // --cpus=0.6
		{"80000 100000", 800},  // --cpus=0.8
		{"100000 100000", 1000}, // --cpus=1.0
		{"110000 100000", 1100}, // --cpus=1.1
	}
	for _, c := range cases {
		read := fakeFS(map[string]string{"/sys/fs/cgroup/cpu.max": c.quotaPeriod})
		got := detectCPUMillisFrom(read)
		if got != c.wantMillis {
			t.Errorf("quota/period %q: got %d millis, want %d", c.quotaPeriod, got, c.wantMillis)
		}
	}
}

func TestDetectCPUMillisFrom_CgroupV2_Unlimited(t *testing.T) {
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/cpu.max": "max 100000\n",
	})
	got := detectCPUMillisFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0 (unlimited)", got)
	}
}

func TestDetectCPUMillisFrom_CgroupV1_QuotaPeriod(t *testing.T) {
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/cpu/cpu.cfs_quota_us":  "50000\n",
		"/sys/fs/cgroup/cpu/cpu.cfs_period_us": "100000\n",
	})
	got := detectCPUMillisFrom(read)
	if got != 500 {
		t.Errorf("got %d, want 500", got)
	}
}

func TestDetectCPUMillisFrom_CgroupV1_UnlimitedQuota(t *testing.T) {
	// cgroup v1 convention: quota -1 means unlimited.
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/cpu/cpu.cfs_quota_us":  "-1\n",
		"/sys/fs/cgroup/cpu/cpu.cfs_period_us": "100000\n",
	})
	got := detectCPUMillisFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0 (unlimited)", got)
	}
}

func TestDetectCPUMillisFrom_NoCgroup(t *testing.T) {
	// Neither v2 nor v1 files present — bare-metal or non-Linux.
	read := fakeFS(map[string]string{})
	got := detectCPUMillisFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0 (no cgroup found)", got)
	}
}

func TestDetectCPUMillisFrom_MalformedContent(t *testing.T) {
	cases := []map[string]string{
		{"/sys/fs/cgroup/cpu.max": ""},
		{"/sys/fs/cgroup/cpu.max": "not-a-number 100000"},
		{"/sys/fs/cgroup/cpu.max": "50000"}, // missing period field
	}
	for _, files := range cases {
		got := detectCPUMillisFrom(fakeFS(files))
		if got != 0 {
			t.Errorf("malformed content %+v: got %d, want 0", files, got)
		}
	}
}

func TestDetectMemoryLimitFrom_CgroupV2(t *testing.T) {
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/memory.max": "536870912\n", // 512 MiB
	})
	got := detectMemoryLimitFrom(read)
	if got != 536870912 {
		t.Errorf("got %d, want 536870912", got)
	}
}

func TestDetectMemoryLimitFrom_CgroupV2_Unlimited(t *testing.T) {
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/memory.max": "max\n",
	})
	got := detectMemoryLimitFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0 (unlimited)", got)
	}
}

func TestDetectMemoryLimitFrom_CgroupV1(t *testing.T) {
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/memory/memory.limit_in_bytes": "536870912\n",
	})
	got := detectMemoryLimitFrom(read)
	if got != 536870912 {
		t.Errorf("got %d, want 536870912", got)
	}
}

func TestDetectMemoryLimitFrom_CgroupV1_UnboundedSentinel(t *testing.T) {
	// cgroup v1's default when no limit is set: a huge, practically-max
	// value (kernel LLONG_MAX rounded to a page boundary), not "-1" or
	// "max". Must be treated as unlimited, not as an 8-exabyte quota.
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/memory/memory.limit_in_bytes": "9223372036854771712\n",
	})
	got := detectMemoryLimitFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0 (unbounded sentinel treated as unlimited)", got)
	}
}

func TestDetectMemoryLimitFrom_NoCgroup(t *testing.T) {
	read := fakeFS(map[string]string{})
	got := detectMemoryLimitFrom(read)
	if got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestDetectMemoryLimitFrom_V2PreferredOverV1(t *testing.T) {
	// When both hierarchies are somehow present, v2 wins (checked first).
	read := fakeFS(map[string]string{
		"/sys/fs/cgroup/memory.max":                   "536870912",
		"/sys/fs/cgroup/memory/memory.limit_in_bytes": "1073741824",
	})
	got := detectMemoryLimitFrom(read)
	if got != 536870912 {
		t.Errorf("got %d, want 536870912 (v2 value)", got)
	}
}

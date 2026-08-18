package capability

import (
	"os"
	"strconv"
	"strings"
)

// cgroupUnlimitedMemoryThreshold is the practical cutoff above which a
// cgroup v1 memory.limit_in_bytes value is treated as "no limit set".
// The kernel's actual sentinel is platform page-size dependent (typically
// near LLONG_MAX rounded down to a page boundary), so we compare against
// a generous 1 EiB threshold rather than an exact constant.
const cgroupUnlimitedMemoryThreshold = int64(1) << 60

// fileReader reads a file's full contents, matching os.ReadFile's
// signature. Detection functions take one as a parameter so tests can
// inject synthetic cgroup file contents without a real cgroup filesystem
// — this machine (Windows dev box) has no /sys/fs/cgroup at all, and CI
// may not run under a CPU/memory-limited container either.
type fileReader func(string) ([]byte, error)

// detectCPUMillis returns the effective CPU limit in milli-cores (1000 =
// one full core) from the current cgroup, or 0 if no limit is set or the
// platform has no cgroups (non-Linux). Callers should fall back to
// cpu_cores * 1000 when this returns 0.
func detectCPUMillis() int32 {
	return detectCPUMillisFrom(os.ReadFile)
}

func detectCPUMillisFrom(read fileReader) int32 {
	// cgroup v2: single unified file "<quota> <period>" in microseconds,
	// or "max <period>" when unlimited.
	if data, err := read("/sys/fs/cgroup/cpu.max"); err == nil {
		if millis, ok := parseCgroupV2CPUMax(string(data)); ok {
			return millis
		}
		return 0 // "max" or unparseable — no limit
	}

	// cgroup v1: quota and period in separate files.
	quotaData, quotaErr := read("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	periodData, periodErr := read("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if quotaErr == nil && periodErr == nil {
		if millis, ok := parseCgroupV1CPUQuota(string(quotaData), string(periodData)); ok {
			return millis
		}
	}

	return 0
}

// parseCgroupV2CPUMax parses cpu.max content ("50000 100000" or
// "max 100000") into milli-cores. Returns ok=false for "max" (unlimited)
// or malformed content.
func parseCgroupV2CPUMax(content string) (int32, bool) {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) != 2 {
		return 0, false
	}
	if fields[0] == "max" {
		return 0, false
	}
	quota, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || quota <= 0 {
		return 0, false
	}
	period, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || period <= 0 {
		return 0, false
	}
	return quotaToMillis(quota, period), true
}

// parseCgroupV1CPUQuota parses the separate cfs_quota_us / cfs_period_us
// files into milli-cores. A quota of -1 means unlimited (cgroup v1
// convention).
func parseCgroupV1CPUQuota(quotaStr, periodStr string) (int32, bool) {
	quota, err := strconv.ParseInt(strings.TrimSpace(quotaStr), 10, 64)
	if err != nil || quota <= 0 {
		return 0, false // -1 (or any non-positive value) means unlimited
	}
	period, err := strconv.ParseInt(strings.TrimSpace(periodStr), 10, 64)
	if err != nil || period <= 0 {
		return 0, false
	}
	return quotaToMillis(quota, period), true
}

// quotaToMillis converts a cfs quota/period pair (both in microseconds)
// to milli-cores: (quota/period) is the fraction of one core, and
// milli-cores = that fraction * 1000.
func quotaToMillis(quota, period int64) int32 {
	millis := quota * 1000 / period
	if millis <= 0 {
		return 0
	}
	if millis > 1<<30 { // guard against pathological/overflowed input
		return 1 << 30
	}
	return int32(millis)
}

// detectMemoryLimit returns the effective memory limit in bytes from the
// current cgroup, or 0 if no limit is set or the platform has no
// cgroups. Callers should fall back to the whole-machine memory reading
// when this returns 0.
func detectMemoryLimit() int64 {
	return detectMemoryLimitFrom(os.ReadFile)
}

func detectMemoryLimitFrom(read fileReader) int64 {
	// cgroup v2: single file, bytes or "max".
	if data, err := read("/sys/fs/cgroup/memory.max"); err == nil {
		return parseCgroupMemoryLimit(string(data))
	}

	// cgroup v1.
	if data, err := read("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		return parseCgroupMemoryLimit(string(data))
	}

	return 0
}

// parseCgroupMemoryLimit parses a memory limit file's content (v1 and v2
// share the same "<bytes>" or "max" format) into bytes, treating both
// the literal "max" sentinel and cgroup v1's practically-unbounded
// default value as "no limit" (0).
func parseCgroupMemoryLimit(content string) int64 {
	trimmed := strings.TrimSpace(content)
	if trimmed == "max" {
		return 0
	}
	limit, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || limit <= 0 {
		return 0
	}
	if limit >= cgroupUnlimitedMemoryThreshold {
		return 0
	}
	return limit
}

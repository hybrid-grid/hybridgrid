// Package taskid generates task and build-session identifiers for
// HybridGrid clients.
//
// Task IDs identify a single unit of work (one translation unit, one
// flutter/unity build). Build session IDs group tasks that belong to one
// logical build: hgbuild runs once per translation unit under make -jN,
// so the only context shared across a real build is the parent make
// environment. A per-invocation session ID would therefore degenerate to
// one "build" per task; see BuildSessionID.
package taskid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"time"
)

// BuildIDEnvVar is the environment variable a wrapping build tool (e.g.
// make) exports to group the per-TU hgbuild invocations of one logical
// build into a single session:
//
//	HG_BUILD_ID=$(uuidgen) make -j5
//
// make exports its environment to every recipe, so each hgbuild process
// inherits the same value and the coordinator can group their tasks.
const BuildIDEnvVar = "HG_BUILD_ID"

// maxIDLength matches validation.MaxTaskIDLength on the coordinator;
// build session IDs share the bound so the same limit applies end to end.
const maxIDLength = 128

var (
	// idRegex is the task-ID charset (alphanumeric, dash, underscore),
	// matching the coordinator's validation.taskIDRegex.
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// buildIDRegex is the build-session charset: the task-ID charset
	// plus dots, because real session labels are versioned
	// (cpython-rel-3.14.0). Dots are safe downstream (JSON value, map
	// key, log field) — they never reach a filesystem path.
	buildIDRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// NewTaskID returns a new unique task identifier.
func NewTaskID() string {
	return "task-" + randomSuffix()
}

// BuildSessionID returns the build session identifier for this client
// process: the HG_BUILD_ID environment value when set and well-formed,
// otherwise a freshly generated ID shared by all tasks this process
// submits (multi-file `hgbuild build a.c b.c` groups under it).
//
// An invalid HG_BUILD_ID (wrong charset or over length) is reported on
// stderr and replaced with a generated ID rather than passed through:
// the coordinator drops malformed grouping keys, and failing the build
// over a cosmetic label would be disproportionate.
func BuildSessionID() string {
	if id := os.Getenv(BuildIDEnvVar); id != "" {
		if buildIDRegex.MatchString(id) && len(id) <= maxIDLength {
			return id
		}
		fmt.Fprintf(os.Stderr, "hgbuild: ignoring invalid %s=%q (must match %s, max %d chars)\n",
			BuildIDEnvVar, id, buildIDRegex.String(), maxIDLength)
	}
	return "build-" + randomSuffix()
}

// randomSuffix is the identifier body shared by both ID kinds: 8 random
// bytes plus a timestamp tail, falling back to a timestamp-only value if
// the system entropy source is unavailable.
func randomSuffix() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%d", hex.EncodeToString(b), time.Now().UnixNano()%10000)
}

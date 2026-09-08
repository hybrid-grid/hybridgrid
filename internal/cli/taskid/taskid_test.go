package taskid

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestNewTaskID_Format(t *testing.T) {
	id := NewTaskID()
	if !strings.HasPrefix(id, "task-") {
		t.Errorf("NewTaskID = %q, want task- prefix", id)
	}
	if !idRegex.MatchString(id) || len(id) > maxIDLength {
		t.Errorf("NewTaskID = %q violates ID charset/length", id)
	}
}

func TestNewTaskID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := range 1000 {
		id := NewTaskID()
		if seen[id] {
			t.Fatalf("duplicate task ID after %d draws: %s", i, id)
		}
		seen[id] = true
	}
}

func TestBuildSessionID_EnvUsedWhenSet(t *testing.T) {
	t.Setenv(BuildIDEnvVar, "cpython-rel-3.14.0")
	if got := BuildSessionID(); got != "cpython-rel-3.14.0" {
		t.Errorf("BuildSessionID = %q, want env value", got)
	}
}

func TestBuildSessionID_FreshWhenUnset(t *testing.T) {
	if id := BuildSessionID(); !strings.HasPrefix(id, "build-") {
		t.Errorf("BuildSessionID = %q, want build- prefix when env unset", id)
	}
}

func TestBuildSessionID_InvalidEnvReplaced(t *testing.T) {
	// Malformed values must not pass through: the coordinator drops
	// them, and the client should group under a valid generated ID.
	for _, bad := range []string{"has spaces", "sla/shes", "unicode-Ω", strings.Repeat("x", maxIDLength+1)} {
		t.Setenv(BuildIDEnvVar, bad)
		got := BuildSessionID()
		if got == bad {
			t.Errorf("BuildSessionID passed invalid env value %q through", bad)
		}
		if !strings.HasPrefix(got, "build-") {
			t.Errorf("BuildSessionID = %q, want generated fallback for invalid %q", got, bad)
		}
	}
}

func TestBuildSessionID_MaxLengthEnvAccepted(t *testing.T) {
	id := strings.Repeat("a", maxIDLength)
	t.Setenv(BuildIDEnvVar, id)
	if got := BuildSessionID(); got != id {
		t.Errorf("BuildSessionID = %q, want max-length env value accepted", got)
	}
}

func TestIDCharset(t *testing.T) {
	// The charset contract both client and coordinator rely on.
	valid := regexp.MustCompile(idRegex.String())
	for _, id := range []string{"a", "task-abc123-42", "build-deadbeef-99", "A_b-C"} {
		if !valid.MatchString(id) {
			t.Errorf("expected %q to match ID charset", id)
		}
	}
	for _, id := range []string{"", "a b", "a.b", "a:b", "a/b"} {
		if valid.MatchString(id) {
			t.Errorf("expected %q to be rejected", id)
		}
	}
}

// TestEnvVarName guards the contract with the wrapping build tool.
func TestEnvVarName(t *testing.T) {
	if BuildIDEnvVar != "HG_BUILD_ID" {
		t.Errorf("BuildIDEnvVar = %q, want HG_BUILD_ID", BuildIDEnvVar)
	}
	_ = os.Getenv(BuildIDEnvVar) // referenced so the os import stays honest
}

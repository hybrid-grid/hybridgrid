// Package builder is the Go SDK for hybridgrid plugin builders
// (builderv1alpha1 exec contract). Implement Builder and call Main
// from your plugin's main(): the SDK speaks the capability/execute
// verbs, decodes the task JSON from stdin, and writes the result JSON
// to stdout. Console text goes to stderr — stdout is reserved for the
// result.
//
// A minimal plugin:
//
//	package main
//
//	import (
//		"context"
//
//		"github.com/h3nr1-d14z/hybridgrid/sdk/builder"
//	)
//
//	type myBuilder struct{}
//
//	func (myBuilder) Capability() builder.Capability {
//		return builder.Capability{ID: "mytool", Description: "builds with mytool"}
//	}
//
//	func (myBuilder) Build(ctx context.Context, t builder.Task) (builder.Result, error) {
//		// run the toolchain against files under t.WorkDir
//		return builder.Result{Success: true, Outputs: []string{"dist/app.bin"}}, nil
//	}
//
//	func main() { builder.Main(myBuilder{}) }
//
// Build it, name the binary hg-exec-mytool, drop it in the worker's
// --builder-dir: the worker discovers and advertises it.
package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// APIVersion must match the worker's builder contract version.
const APIVersion = "builderv1alpha1"

// Capability identifies the plugin in the capability handshake.
type Capability struct {
	ID          string `json:"id"`
	API         string `json:"api"`
	Description string `json:"description,omitempty"`
}

// Task is the build request piped to the plugin's stdin. WorkDir holds
// the extracted source archive; Outputs in Result are WorkDir-relative.
type Task struct {
	TaskID    string            `json:"task_id"`
	BuildID   string            `json:"build_id,omitempty"`
	BuilderID string            `json:"builder_id"`
	WorkDir   string            `json:"work_dir"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	TimeoutMs int64             `json:"timeout_ms,omitempty"`
}

// Result is the plugin's structured answer, written to stdout on
// exit code 0. Outputs are WorkDir-relative paths (files only) that
// the worker archives back to the client.
type Result struct {
	Success bool     `json:"success"`
	Outputs []string `json:"outputs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Builder is what a plugin implements.
type Builder interface {
	Capability() Capability
	Build(ctx context.Context, task Task) (Result, error)
}

// Main runs the plugin protocol for b and exits the process. Call it
// from the plugin's main().
func Main(b Builder) {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: hg-exec-<id> capability|execute")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "capability":
		cap := b.Capability()
		cap.API = APIVersion
		if cap.ID == "" {
			fmt.Fprintln(os.Stderr, "builder: Capability().ID must not be empty")
			os.Exit(2)
		}
		if err := json.NewEncoder(os.Stdout).Encode(cap); err != nil {
			fmt.Fprintln(os.Stderr, "builder: encode capability:", err)
			os.Exit(2)
		}

	case "execute":
		var task Task
		if err := json.NewDecoder(os.Stdin).Decode(&task); err != nil {
			fmt.Fprintln(os.Stderr, "builder: decode task:", err)
			os.Exit(2)
		}
		ctx := context.Background()
		result, err := b.Build(ctx, task)
		if err != nil {
			result = Result{Success: false, Error: err.Error()}
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, "builder: encode result:", err)
			os.Exit(2)
		}
		if !result.Success {
			// A non-zero exit keeps failure unmistakable even for
			// hosts that ignore stdout.
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "builder: unknown verb %q\n", os.Args[1])
		os.Exit(2)
	}
}

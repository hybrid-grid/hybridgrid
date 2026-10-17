// hg-exec-script is the reference plugin builder: it runs the
// project's own build script (hg-build.sh, build.sh, or build.bat)
// inside the extracted source, so any toolchain becomes a distributed
// hybridgrid build without writing a line of plugin Go. Pass-through
// args go to the script. Outputs are read from a .hg-outputs file
// (one WorkDir-relative path per line) when the script writes one;
// otherwise every file under dist/ is collected.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/h3nr1-d14z/hybridgrid/sdk/builder"
)

const (
	outputsFile = ".hg-outputs"
	defaultOut  = "dist"
)

type scriptBuilder struct{}

func (scriptBuilder) Capability() builder.Capability {
	return builder.Capability{
		ID:          "script",
		Description: "runs the project's hg-build.sh/build.sh; any toolchain, zero plugin code",
	}
}

func (scriptBuilder) Build(ctx context.Context, task builder.Task) (builder.Result, error) {
	script, err := findScript(task.WorkDir)
	if err != nil {
		return builder.Result{}, err
	}

	// The script inherits the plugin's stderr (the console) and its
	// stdout too: the SDK reserves only the plugin's own stdout for
	// the result JSON, which is written after the script finishes.
	cmd := exec.CommandContext(ctx, script, task.Args...)
	cmd.Dir = task.WorkDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), envSlice(task.Env)...)

	fmt.Fprintf(os.Stderr, "script: running %s %v\n", filepath.Base(script), task.Args)
	if err := cmd.Run(); err != nil {
		return builder.Result{Success: false, Error: fmt.Sprintf("script failed: %v", err)}, nil
	}

	outputs, err := collectOutputs(task.WorkDir)
	if err != nil {
		return builder.Result{}, err
	}
	if len(outputs) == 0 {
		return builder.Result{Success: false, Error: "script succeeded but produced no outputs (write .hg-outputs or build into dist/)"}, nil
	}
	return builder.Result{Success: true, Outputs: outputs}, nil
}

// findScript locates the project's build script.
func findScript(dir string) (string, error) {
	names := []string{"hg-build.sh", "build.sh"}
	if runtime.GOOS == "windows" {
		names = []string{"hg-build.bat", "build.bat", "hg-build.ps1", "build.ps1"}
	}
	for _, name := range names {
		p := filepath.Join(dir, name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("no build script found (tried %v)", names)
}

// collectOutputs reads .hg-outputs when present, else walks dist/.
func collectOutputs(workDir string) ([]string, error) {
	manifest := filepath.Join(workDir, outputsFile)
	if info, err := os.Stat(manifest); err == nil && !info.IsDir() {
		file, err := os.Open(manifest)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", outputsFile, err)
		}
		defer file.Close()
		var outputs []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			outputs = append(outputs, filepath.ToSlash(filepath.Clean(line)))
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return outputs, nil
	}

	var outputs []string
	dist := filepath.Join(workDir, defaultOut)
	err := filepath.WalkDir(dist, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if entry == nil && path == dist {
				return nil // no dist/ at all
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(workDir, path)
		if err != nil {
			return err
		}
		outputs = append(outputs, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return outputs, nil
}

// envSlice flattens the task env map into KEY=VALUE strings.
func envSlice(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

func main() {
	builder.Main(scriptBuilder{})
}

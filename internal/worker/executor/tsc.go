package executor

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

// TscExecutor type-checks and emits one TypeScript project with
// `tsc -p`. `hgbuild tsc -b` splits a project-references build into one
// such task per project, sending each only after the projects it
// references have been built and their declarations are in its source.
//
// The compiler is the project's own node_modules/typescript when the
// dependency layer carries it, so every machine type-checks with the
// version the project pinned; tsc on PATH is the fallback.
type TscExecutor struct {
	node string // node binary, for node_modules/typescript/bin/tsc
	tsc  string // fallback tsc on PATH
}

// NewTscExecutor returns an executor using node and tsc from PATH.
func NewTscExecutor() *TscExecutor {
	return &TscExecutor{node: "node", tsc: "tsc"}
}

func (e *TscExecutor) Name() string { return "tsc" }

func (e *TscExecutor) CanExecute(pb.Architecture, pb.Architecture) bool { return true }

func (e *TscExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	project, err := tscProject(req.NodeConfig)
	if err != nil {
		return nil, err
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-tsc-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	if req.OpenDeps != nil {
		if err := extractDeps(req.OpenDeps, workDir); err != nil {
			return nil, err
		}
	}
	if err := extractSourceArchive(workDir, req.SourceArchive); err != nil {
		return nil, fmt.Errorf("failed to extract source: %w", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, project)); err != nil {
		return nil, fmt.Errorf("tsc project %s not in the source archive", project)
	}

	before, err := snapshotFiles(workDir)
	if err != nil {
		return nil, err
	}

	argv := e.command(workDir, project)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workDir
	stdout := newLimitedBuffer(maxFlutterLogBytes)
	stderr := newLimitedBuffer(maxFlutterLogBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	runErr := cmd.Run()
	result := &Result{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		CompilationTime: time.Since(start),
	}
	if runErr != nil {
		exitErr, ok := runErr.(*exec.ExitError)
		if !ok {
			return nil, fmt.Errorf("failed to run %s: %w", argv[0], runErr)
		}
		// tsc reports type errors on stdout and exits non-zero.
		result.ExitCode = int32(exitErr.ExitCode())
		return result, nil
	}

	outputs, err := changedFiles(workDir, before)
	if err != nil {
		return nil, err
	}
	entries := make([]flutterArtifact, 0, len(outputs))
	for _, rel := range outputs {
		abs := filepath.Join(workDir, filepath.FromSlash(rel))
		sum, size, err := checksumFile(abs)
		if err != nil {
			return nil, fmt.Errorf("failed to checksum %s: %w", rel, err)
		}
		entries = append(entries, flutterArtifact{
			info:    &pb.ArtifactInfo{Name: filepath.Base(rel), Path: rel, SizeBytes: size, Checksum: sum},
			absPath: abs,
		})
	}
	archive, err := archiveArtifacts(workDir, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to archive outputs: %w", err)
	}
	result.Success = true
	result.ArtifactArchive = archive
	for _, en := range entries {
		result.Artifacts = append(result.Artifacts, en.info)
	}
	return result, nil
}

// tscProject validates the project path: relative, inside the archive.
func tscProject(cfg *pb.NodeConfig) (string, error) {
	p := strings.TrimSpace(cfg.GetTscProject())
	if p == "" {
		return "", fmt.Errorf("node_config.tsc_project required")
	}
	clean := filepath.Clean(filepath.FromSlash(p))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid tsc_project %q", p)
	}
	return clean, nil
}

// command prefers the nearest node_modules/typescript above the project.
func (e *TscExecutor) command(workDir, project string) []string {
	dir := project
	if info, err := os.Stat(filepath.Join(workDir, project)); err == nil && !info.IsDir() {
		dir = filepath.Dir(project)
	}
	for {
		bin := filepath.Join(workDir, dir, "node_modules", "typescript", "bin", "tsc")
		if _, err := os.Stat(bin); err == nil {
			return []string{e.node, bin, "-p", project}
		}
		if dir == "." || dir == string(filepath.Separator) {
			break
		}
		dir = filepath.Dir(dir)
	}
	return []string{e.tsc, "-p", project}
}

func extractDeps(open func() (io.ReadCloser, error), dest string) error {
	rc, err := open()
	if err != nil {
		return fmt.Errorf("failed to open dependency layer: %w", err)
	}
	defer rc.Close()
	if err := depslayer.Extract(rc, dest); err != nil {
		return fmt.Errorf("failed to extract dependency layer: %w", err)
	}
	return nil
}

type fileStamp struct {
	size  int64
	mtime time.Time
}

// snapshotFiles records every regular file outside node_modules.
func snapshotFiles(root string) (map[string]fileStamp, error) {
	out := make(map[string]fileStamp)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = fileStamp{size: info.Size(), mtime: info.ModTime()}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan work dir: %w", err)
	}
	return out, nil
}

// changedFiles lists files created or rewritten since before: what the
// compiler emitted (.js, .d.ts, maps, .tsbuildinfo), wherever outDir,
// declarationDir or the default next-to-source layout put them.
func changedFiles(root string, before map[string]fileStamp) ([]string, error) {
	after, err := snapshotFiles(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for rel, st := range after {
		if old, ok := before[rel]; !ok || old != st {
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out, nil
}

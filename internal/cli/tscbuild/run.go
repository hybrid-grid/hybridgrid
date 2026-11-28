package tscbuild

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/sourcearchive"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/taskid"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

// BuildClient is the coordinator API the build needs.
type BuildClient interface {
	Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error)
}

// Result is the outcome of one project.
type Result struct {
	Project *Project
	Worker  string
	Took    time.Duration
	Files   int
	// Output is the compiler's diagnostics (tsc writes them to stdout).
	Output string
	Err    error
	// Skipped means a referenced project failed, so this one never ran.
	Skipped bool
}

// Options tunes a run.
type Options struct {
	BuildID        string
	TimeoutSeconds int32
	// OnDone, when set, is called as each project finishes.
	OnDone func(Result)
}

// deps is the run's dependency layer, archived only if a receiver
// reports it missing.
type deps struct {
	layer *depslayer.Layer
	once  sync.Once
	data  []byte
	err   error
}

func (d *deps) archive() ([]byte, error) {
	d.once.Do(func() { d.data, d.err = d.layer.Archive() })
	return d.data, d.err
}

// Run builds every project of g, each as soon as its references are
// built, and writes their outputs back under g.Root. Results come back
// in g.Projects order.
func Run(ctx context.Context, client BuildClient, g *Graph, opts Options) ([]Result, error) {
	var layer *deps
	switch l, err := depslayer.Scan(g.Root); {
	case err == nil:
		layer = &deps{layer: l}
	case !errors.Is(err, depslayer.ErrNoDeps):
		return nil, err
	}
	if opts.BuildID == "" {
		opts.BuildID = taskid.BuildSessionID()
	}

	done := make(map[string]chan struct{}, len(g.Projects))
	results := make(map[string]*Result, len(g.Projects))
	for _, p := range g.Projects {
		done[p.Config] = make(chan struct{})
		results[p.Config] = &Result{Project: p}
	}
	// Outputs are extracted into g.Root, which later tasks archive, so
	// extraction and archiving must not interleave.
	var rootMu sync.RWMutex

	var wg sync.WaitGroup
	for _, p := range g.Projects {
		wg.Add(1)
		go func(p *Project) {
			defer wg.Done()
			defer close(done[p.Config])
			res := results[p.Config]
			defer func() {
				if opts.OnDone != nil {
					opts.OnDone(*res)
				}
			}()
			for _, ref := range p.Refs {
				select {
				case <-done[ref]:
				case <-ctx.Done():
					res.Err = ctx.Err()
					return
				}
				if r := results[ref]; r.Err != nil || r.Skipped {
					res.Skipped = true
					return
				}
			}
			start := time.Now()
			rootMu.RLock()
			archive, hash, err := sourcearchive.Create(g.Root, g.excludeFor(p))
			rootMu.RUnlock()
			if err != nil {
				res.Err = err
				return
			}
			resp, err := buildOne(ctx, client, layer, &pb.BuildRequest{
				TaskId:         taskid.NewTaskID(),
				BuildId:        opts.BuildID,
				BuildType:      pb.BuildType_BUILD_TYPE_NODEJS,
				SourceHash:     hash,
				SourceArchive:  archive,
				Config:         &pb.BuildRequest_NodeConfig{NodeConfig: &pb.NodeConfig{TscProject: p.Config}},
				TimeoutSeconds: opts.TimeoutSeconds,
			})
			res.Took = time.Since(start)
			if err != nil {
				res.Err = err
				return
			}
			res.Worker = resp.GetWorkerId()
			res.Output = strings.TrimSpace(resp.GetStdout() + "\n" + resp.GetStderr())
			if resp.GetStatus() != pb.TaskStatus_STATUS_COMPLETED {
				res.Err = fmt.Errorf("tsc failed for %s (exit %d)", p.Config, resp.GetExitCode())
				return
			}
			rootMu.Lock()
			res.Files, err = extract(resp.GetArtifacts(), g.Root)
			rootMu.Unlock()
			if err != nil {
				res.Err = err
			}
		}(p)
	}
	wg.Wait()

	out := make([]Result, 0, len(g.Projects))
	for _, p := range g.Projects {
		out = append(out, *results[p.Config])
	}
	return out, nil
}

// buildOne sends the request naming the dependency layer by hash, and
// resends it with the layer attached if the coordinator lacks it.
func buildOne(ctx context.Context, client BuildClient, layer *deps, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	if layer != nil {
		req.DepsHash = layer.layer.Hash
	}
	resp, err := client.Build(ctx, req)
	if layer == nil || !isMissingDeps(err) {
		return resp, err
	}
	data, aerr := layer.archive()
	if aerr != nil {
		return nil, fmt.Errorf("failed to archive node_modules: %w", aerr)
	}
	req.DepsArchive = data
	return client.Build(ctx, req)
}

func isMissingDeps(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.FailedPrecondition && st.Message() == depslayer.MissingMessage
}

// excludeFor keeps everything a `tsc -p` of p can read: the whole tree
// except VCS data, node_modules (sent as the dependency layer) and the
// directories of projects p does not reference.
func (g *Graph) excludeFor(p *Project) sourcearchive.ExcludeFunc {
	closure := g.Closure(p)
	dirs := g.projectDirs(closure)
	var needed []string
	for dir, in := range dirs {
		if in {
			needed = append(needed, dir)
		}
	}
	return func(rel string, isDir bool) bool {
		base := rel[strings.LastIndex(rel, "/")+1:]
		switch base {
		case ".git", ".hg", ".svn", "node_modules":
			return true
		}
		if !isDir {
			return false
		}
		in, isProject := dirs[rel]
		if !isProject || in {
			return false
		}
		// Keep an unrelated project's directory when a needed project
		// lives inside it.
		for _, n := range needed {
			if n != "." && strings.HasPrefix(n+"/", rel+"/") {
				return false
			}
		}
		return true
	}
}

// extract writes the regular files of a task's output archive under
// root, refusing paths that escape it, and returns how many it wrote.
func extract(archive []byte, root string) (int, error) {
	if len(archive) == 0 {
		return 0, nil
	}
	tr := tar.NewReader(bytes.NewReader(archive))
	n := 0
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return n, nil
		}
		if err != nil {
			return n, fmt.Errorf("failed to read outputs: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		clean := filepath.Clean(filepath.FromSlash(hdr.Name))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return n, fmt.Errorf("output %q escapes the project", hdr.Name)
		}
		dest := filepath.Join(root, clean)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return n, err
		}
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return n, err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return n, err
		}
		if err := f.Close(); err != nil {
			return n, err
		}
		n++
	}
}

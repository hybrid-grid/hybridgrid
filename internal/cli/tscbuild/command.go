package tscbuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client is a BuildClient that must be closed.
type Client interface {
	BuildClient
	Close() error
}

// Dependencies wires the command to the coordinator.
type Dependencies struct {
	CoordinatorAddr func() string
	NewClient       func(addr string, timeout time.Duration) (Client, error)
	RequestTimeout  time.Duration
	// NoFallback reports whether --no-fallback is set; otherwise an
	// unreachable coordinator falls back to a local `tsc -b`.
	NoFallback func() bool
	// LocalBuild runs the local fallback (default: tsc -b <path>).
	LocalBuild func(ctx context.Context, path string, stdout, stderr io.Writer) error
}

// NewCommand returns `hgbuild tsc`.
func NewCommand(deps Dependencies) *cobra.Command {
	var build bool
	cmd := &cobra.Command{
		Use:   "tsc [-b] [project]",
		Short: "Build a TypeScript project-references graph across workers",
		Long: `Builds a TypeScript project-references graph (what tsc -b builds) on the
cluster: every referenced project becomes one tsc -p task, sent as soon as
the projects it references are built, so independent projects compile in
parallel on different machines. Outputs (.js, .d.ts, .tsbuildinfo) are
written back exactly where tsc -b would write them.

node_modules is sent once per change as a separate layer and cached by the
coordinator and each worker; workers run the project's own
node_modules/typescript when present.`,
		Example: `  hgbuild tsc -b                 # tsconfig.json in the current directory
  hgbuild tsc -b packages/app    # a project and everything it references`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		// hgbuild's main prints the returned error once.
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			return run(cmd, deps, path)
		},
	}
	cmd.Flags().BoolVarP(&build, "build", "b", true, "build mode (follows project references); always on, accepted for tsc -b familiarity")
	return cmd
}

func run(cmd *cobra.Command, deps Dependencies, path string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	g, err := Load(path)
	if err != nil {
		return err
	}

	fallback := func(reason error) error {
		if deps.NoFallback != nil && deps.NoFallback() {
			return reason
		}
		fmt.Fprintf(out, "Coordinator unavailable (%v); building locally with tsc -b\n", reason)
		local := deps.LocalBuild
		if local == nil {
			local = localTsc
		}
		return local(ctx, path, out, cmd.ErrOrStderr())
	}

	client, err := connect(deps)
	if err != nil {
		return fallback(err)
	}
	defer client.Close()

	fmt.Fprintf(out, "Building %d TypeScript projects on the cluster\n", len(g.Projects))
	start := time.Now()
	var (
		mu          sync.Mutex
		unavailable bool
	)
	results, err := Run(ctx, client, g, Options{OnDone: func(r Result) {
		mu.Lock()
		defer mu.Unlock()
		// An unreachable coordinator is reported once, by the fallback or
		// the final error, not per project.
		if status.Code(r.Err) == codes.Unavailable {
			unavailable = true
			return
		}
		if r.Skipped && unavailable {
			return
		}
		printResult(out, r)
	}})
	if err != nil {
		return err
	}
	if err := unreachable(results); err != nil {
		return fallback(err)
	}

	var summed time.Duration
	var failed, skipped []string
	for _, r := range results {
		summed += r.Took
		switch {
		case r.Skipped:
			skipped = append(skipped, r.Project.Config)
		case r.Err != nil:
			failed = append(failed, r.Project.Config)
		}
	}
	for _, r := range results {
		if r.Err != nil && r.Output != "" {
			fmt.Fprintf(out, "\n%s:\n%s\n", r.Project.Config, r.Output)
		}
	}
	if len(failed) > 0 {
		msg := fmt.Sprintf("%d of %d projects failed: %s", len(failed), len(results), strings.Join(failed, ", "))
		if len(skipped) > 0 {
			msg += fmt.Sprintf(" (%d skipped: %s)", len(skipped), strings.Join(skipped, ", "))
		}
		return fmt.Errorf("%s", msg)
	}
	fmt.Fprintf(out, "Done in %.1fs (tasks summed %.1fs)\n", time.Since(start).Seconds(), summed.Seconds())
	return nil
}

func printResult(out io.Writer, r Result) {
	switch {
	case r.Skipped:
		fmt.Fprintf(out, "  %-40s skipped (a referenced project failed)\n", r.Project.Config)
	case r.Err != nil:
		fmt.Fprintf(out, "  %-40s %-22s %6.1fs  FAILED: %v\n", r.Project.Config, r.Worker, r.Took.Seconds(), r.Err)
	default:
		fmt.Fprintf(out, "  %-40s %-22s %6.1fs  %d files\n", r.Project.Config, r.Worker, r.Took.Seconds(), r.Files)
	}
}

// unreachable returns the error when no project reached the
// coordinator: every one that ran failed with Unavailable.
func unreachable(results []Result) error {
	var first error
	for _, r := range results {
		if r.Skipped {
			continue
		}
		if status.Code(r.Err) != codes.Unavailable {
			return nil
		}
		if first == nil {
			first = r.Err
		}
	}
	return first
}

func connect(deps Dependencies) (Client, error) {
	if deps.CoordinatorAddr == nil || deps.NewClient == nil {
		return nil, fmt.Errorf("coordinator client not configured")
	}
	addr := strings.TrimSpace(deps.CoordinatorAddr())
	if addr == "" {
		return nil, fmt.Errorf("no coordinator found")
	}
	c, err := deps.NewClient(addr, deps.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	return c, nil
}

func localTsc(ctx context.Context, path string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "tsc", "-b", path)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = stdout, stderr, os.Stdin
	return cmd.Run()
}

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/fallback"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
)

// Service handles distributed compilation with preprocessing and caching.
type Service struct {
	cache        *cache.Store
	client       *client.Client
	fallback     *fallback.LocalFallback
	preprocessor *compiler.Preprocessor
	verbose      bool
	maxRetries   int
	retryDelay   time.Duration
}

// Config holds build service configuration.
type Config struct {
	CacheDir        string
	CacheMaxSize    int64
	CacheTTLHours   int
	CoordinatorAddr string
	Insecure        bool
	Timeout         time.Duration
	FallbackEnabled bool
	Verbose         bool
	MaxRetries      int           // Max retries for transient failures
	RetryDelay      time.Duration // Initial delay between retries
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		CacheDir:        filepath.Join(home, ".hybridgrid", "cache"),
		CacheMaxSize:    10 * 1024 * 1024 * 1024, // 10GB
		CacheTTLHours:   168,                     // 1 week
		CoordinatorAddr: "localhost:9000",
		Insecure:        true,
		Timeout:         5 * time.Minute,
		FallbackEnabled: true,
		Verbose:         false,
		MaxRetries:      3,
		RetryDelay:      100 * time.Millisecond,
	}
}

// New creates a new build service.
func New(cfg Config) (*Service, error) {
	// Initialize cache
	cacheStore, err := cache.NewStore(cfg.CacheDir, cfg.CacheMaxSize, cfg.CacheTTLHours)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize cache, continuing without cache")
		cacheStore = nil
	}

	// Initialize preprocessor
	preprocessor := compiler.NewPreprocessor(compiler.DefaultPreprocessorConfig())

	// Initialize local fallback
	fb := fallback.New(fallback.Config{
		Enabled:    cfg.FallbackEnabled,
		MaxTimeout: cfg.Timeout,
	})

	return &Service{
		cache:        cacheStore,
		preprocessor: preprocessor,
		fallback:     fb,
		verbose:      cfg.Verbose,
		maxRetries:   cfg.MaxRetries,
		retryDelay:   cfg.RetryDelay,
	}, nil
}

// SetClient sets the gRPC client for remote compilation.
func (s *Service) SetClient(c *client.Client) {
	s.client = c
}

// Close closes the build service.
func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Request represents a build request.
type Request struct {
	TaskID     string
	SourceFile string
	OutputFile string
	Args       *compiler.ParsedArgs
	TargetArch pb.Architecture
	Timeout    time.Duration
}

// Result represents a build result.
type Result struct {
	ObjectFile      []byte
	ExitCode        int
	Stdout          string
	Stderr          string
	CacheHit        bool
	Fallback        bool
	FallbackReason  string
	Duration        time.Duration
	PreprocessTime  time.Duration
	CompilationTime time.Duration
	WorkerID        string
}

// Build compiles a source file using the distributed build system.
func (s *Service) Build(ctx context.Context, req *Request) (*Result, error) {
	startTime := time.Now()
	result := &Result{}

	// Step 1: Read raw source file (for cross-compilation support)
	rawSource, err := os.ReadFile(req.SourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	// Collect project include files from -I paths
	includeFiles, err := s.collectIncludeFiles(req.Args)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to collect include files, continuing without them")
		includeFiles = nil
	}

	if s.verbose {
		log.Debug().
			Str("file", req.SourceFile).
			Int("source_size", len(rawSource)).
			Int("include_files", len(includeFiles)).
			Msg("Source loaded for cross-compilation")
	}

	// Step 2: Check cache (using raw source hash)
	cacheKey := s.generateCacheKeyRaw(req, rawSource)
	if s.cache != nil {
		if cached, ok := s.cache.GetBytes(cacheKey); ok {
			result.ObjectFile = cached
			result.CacheHit = true
			result.ExitCode = 0
			result.Duration = time.Since(startTime)

			// Report cache hit to coordinator (for dashboard stats)
			// Must be synchronous because process exits immediately after return
			if s.client != nil {
				_ = s.client.ReportCacheHit(ctx, 1)
			}

			if s.verbose {
				log.Info().
					Str("file", req.SourceFile).
					Str("cache_key", cacheKey).
					Msg("[cache] Cache hit")
			}
			return result, nil
		}
	}

	// Step 3: Try remote compilation (with raw source for cross-compilation)
	if s.client != nil {
		compileResult, err := s.compileRemoteRaw(ctx, req, rawSource, includeFiles)
		if err == nil && compileResult.ExitCode == 0 {
			result.ObjectFile = compileResult.ObjectFile
			result.ExitCode = compileResult.ExitCode
			result.Stdout = compileResult.Stdout
			result.Stderr = compileResult.Stderr
			result.CompilationTime = compileResult.CompilationTime
			result.WorkerID = compileResult.WorkerID
			result.Duration = time.Since(startTime)

			// Store in cache on success
			if s.cache != nil && len(result.ObjectFile) > 0 {
				if err := s.cache.PutBytes(cacheKey, result.ObjectFile); err != nil {
					log.Warn().Err(err).Msg("Failed to store in cache")
				}
			}

			if s.verbose {
				log.Info().
					Str("file", req.SourceFile).
					Str("worker", result.WorkerID).
					Dur("compile_time", result.CompilationTime).
					Msg("[remote] Compilation complete")
			}
			return result, nil
		}

		// Remote failed, log warning and try fallback
		if err != nil {
			log.Warn().Err(err).Str("file", req.SourceFile).Msg("Remote compilation failed, trying fallback")
			result.FallbackReason = fmt.Sprintf("remote error: %v", err)
		} else {
			log.Warn().Int("exit_code", compileResult.ExitCode).Str("stderr", compileResult.Stderr).Msg("Remote compilation returned non-zero exit code")
			// If remote compilation failed with exit code, return that error (not a fallback case)
			result.ObjectFile = compileResult.ObjectFile
			result.ExitCode = compileResult.ExitCode
			result.Stdout = compileResult.Stdout
			result.Stderr = compileResult.Stderr
			result.Duration = time.Since(startTime)
			return result, nil
		}
	} else {
		result.FallbackReason = "no coordinator connection"
	}

	// Step 4: Local fallback (needs preprocessing for local compilation)
	if !s.fallback.IsEnabled() {
		return nil, fmt.Errorf("remote compilation failed and local fallback is disabled")
	}

	// For local fallback, we need to preprocess first
	prepResult, err := s.preprocessor.Preprocess(ctx, req.Args, req.SourceFile)
	if err != nil {
		return nil, fmt.Errorf("preprocessing for fallback failed: %w", err)
	}

	fallbackResult, err := s.compileLocal(ctx, req, prepResult.PreprocessedSource)
	if err != nil {
		return nil, fmt.Errorf("local fallback failed: %w", err)
	}

	result.ObjectFile = fallbackResult.ObjectCode
	result.ExitCode = fallbackResult.ExitCode
	result.Stdout = fallbackResult.Stdout
	result.Stderr = fallbackResult.Stderr
	result.Fallback = true
	result.CompilationTime = fallbackResult.CompilationTime
	result.Duration = time.Since(startTime)

	// Store in cache on success
	if s.cache != nil && result.ExitCode == 0 && len(result.ObjectFile) > 0 {
		if err := s.cache.PutBytes(cacheKey, result.ObjectFile); err != nil {
			log.Warn().Err(err).Msg("Failed to store in cache")
		}
	}

	if s.verbose {
		log.Info().
			Str("file", req.SourceFile).
			Dur("compile_time", result.CompilationTime).
			Str("reason", result.FallbackReason).
			Msg("[local] Fallback compilation complete")
	}

	return result, nil
}

// compileRemote sends compilation to a remote worker with retry logic.
func (s *Service) compileRemote(ctx context.Context, req *Request, preprocessed []byte) (*remoteResult, error) {
	compileReq := &pb.CompileRequest{
		TaskId:             req.TaskID,
		Compiler:           req.Args.Compiler,
		CompilerArgs:       s.buildRemoteArgs(req.Args),
		PreprocessedSource: preprocessed,
		TargetArch:         req.TargetArch,
		TimeoutSeconds:     int32(req.Timeout.Seconds()),
		ClientOs:           getClientOS(),
		ClientArch:         getClientArch(),
	}

	var lastErr error
	delay := s.retryDelay
	maxRetries := s.maxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			if s.verbose {
				log.Debug().
					Int("attempt", attempt+1).
					Int("max_retries", maxRetries).
					Dur("delay", delay).
					Str("task_id", req.TaskID).
					Msg("Retrying remote compilation")
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2 // Exponential backoff
			if delay > 5*time.Second {
				delay = 5 * time.Second // Cap at 5 seconds
			}
		}

		resp, err := s.client.Compile(ctx, compileReq)
		if err != nil {
			lastErr = err
			// Check if error is retryable (network errors, timeouts)
			if isRetryableError(err) {
				continue
			}
			// Non-retryable error, return immediately
			return nil, err
		}

		return &remoteResult{
			ObjectFile:      resp.ObjectFile,
			ExitCode:        int(resp.ExitCode),
			Stdout:          resp.Stdout,
			Stderr:          resp.Stderr,
			CompilationTime: time.Duration(resp.CompilationTimeMs) * time.Millisecond,
			WorkerID:        resp.WorkerId,
		}, nil
	}

	return nil, fmt.Errorf("remote compilation failed after %d attempts: %w", maxRetries, lastErr)
}

// isRetryableError checks if an error is transient and worth retrying.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Retry on network/timeout errors
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") ||
		strings.Contains(errStr, "unavailable") ||
		strings.Contains(errStr, "reset") ||
		strings.Contains(errStr, "EOF")
}

type remoteResult struct {
	ObjectFile      []byte
	ExitCode        int
	Stdout          string
	Stderr          string
	CompilationTime time.Duration
	WorkerID        string
}

// buildRemoteArgs builds compiler arguments for remote compilation.
func (s *Service) buildRemoteArgs(args *compiler.ParsedArgs) []string {
	var remoteArgs []string

	// Add compile-only flag
	remoteArgs = append(remoteArgs, "-c")

	// Add optimization and other flags (but not -I, -D since source is preprocessed)
	for _, flag := range args.Flags {
		// Skip flags that were only needed for preprocessing
		if strings.HasPrefix(flag, "-I") || strings.HasPrefix(flag, "-D") {
			continue
		}
		remoteArgs = append(remoteArgs, flag)
	}

	// Add standard if specified
	if args.Standard != "" {
		remoteArgs = append(remoteArgs, "-std="+args.Standard)
	}

	return remoteArgs
}

// compileLocal compiles using local fallback.
func (s *Service) compileLocal(ctx context.Context, req *Request, preprocessed []byte) (*fallback.CompileResult, error) {
	job := &fallback.CompileJob{
		TaskID:             req.TaskID,
		Compiler:           req.Args.Compiler,
		Args:               s.buildRemoteArgs(req.Args),
		PreprocessedSource: preprocessed,
		Timeout:            req.Timeout,
	}

	return s.fallback.Execute(ctx, job)
}

// generateCacheKey creates a cache key for the compilation.
func (s *Service) generateCacheKey(req *Request, preprocessed []byte) string {
	key := &cache.CompilationKey{
		Compiler:   req.Args.Compiler,
		TargetArch: req.TargetArch.String(),
		Flags:      req.Args.Flags,
		Defines:    req.Args.Defines,
		SourceHash: cache.HashBytes(preprocessed),
	}
	return key.Build()
}

// IsDistributable checks if the compilation can be distributed.
func IsDistributable(args *compiler.ParsedArgs) bool {
	return args.IsDistributable()
}

// getClientOS returns the current operating system name.
func getClientOS() string {
	return runtime.GOOS
}

// getClientArch returns the current architecture as protobuf enum.
func getClientArch() pb.Architecture {
	switch runtime.GOARCH {
	case "amd64":
		return pb.Architecture_ARCH_X86_64
	case "arm64":
		return pb.Architecture_ARCH_ARM64
	case "arm":
		return pb.Architecture_ARCH_ARMV7
	default:
		return pb.Architecture_ARCH_UNSPECIFIED
	}
}

// compileRemoteRaw sends raw source to a remote worker for cross-compilation.
func (s *Service) compileRemoteRaw(ctx context.Context, req *Request, rawSource []byte, includeFiles map[string][]byte) (*remoteResult, error) {
	compileReq := &pb.CompileRequest{
		TaskId:         req.TaskID,
		Compiler:       req.Args.Compiler,
		CompilerArgs:   s.buildRemoteArgsForRaw(req.Args),
		RawSource:      rawSource,
		SourceFilename: filepath.Base(req.SourceFile),
		IncludeFiles:   includeFiles,
		TargetArch:     req.TargetArch,
		TimeoutSeconds: int32(req.Timeout.Seconds()),
		ClientOs:       getClientOS(),
		ClientArch:     getClientArch(),
	}

	var lastErr error
	delay := s.retryDelay
	maxRetries := s.maxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			if s.verbose {
				log.Debug().
					Int("attempt", attempt+1).
					Int("max_retries", maxRetries).
					Dur("delay", delay).
					Str("task_id", req.TaskID).
					Msg("Retrying remote compilation")
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}

		resp, err := s.client.Compile(ctx, compileReq)
		if err != nil {
			lastErr = err
			if isRetryableError(err) {
				continue
			}
			return nil, err
		}

		return &remoteResult{
			ObjectFile:      resp.ObjectFile,
			ExitCode:        int(resp.ExitCode),
			Stdout:          resp.Stdout,
			Stderr:          resp.Stderr,
			CompilationTime: time.Duration(resp.CompilationTimeMs) * time.Millisecond,
			WorkerID:        resp.WorkerId,
		}, nil
	}

	return nil, fmt.Errorf("remote compilation failed after %d attempts: %w", maxRetries, lastErr)
}

// buildRemoteArgsForRaw builds compiler arguments for raw source mode.
// Keeps -I, -D flags since source is not preprocessed.
func (s *Service) buildRemoteArgsForRaw(args *compiler.ParsedArgs) []string {
	var remoteArgs []string

	// Add compile-only flag
	remoteArgs = append(remoteArgs, "-c")

	// Keep all flags including -I and -D (needed for raw source compilation)
	for _, flag := range args.Flags {
		remoteArgs = append(remoteArgs, flag)
	}

	// Add standard if specified
	if args.Standard != "" {
		remoteArgs = append(remoteArgs, "-std="+args.Standard)
	}

	return remoteArgs
}

// collectIncludeFiles collects header files from project -I paths.
// Only includes project-local headers, not system headers.
func (s *Service) collectIncludeFiles(args *compiler.ParsedArgs) (map[string][]byte, error) {
	includeFiles := make(map[string][]byte)

	for _, flag := range args.Flags {
		if !strings.HasPrefix(flag, "-I") {
			continue
		}

		var includePath string
		if flag == "-I" {
			continue // Skip if path is separate (we'd need to track next arg)
		}
		includePath = strings.TrimPrefix(flag, "-I")

		// Skip system paths
		if strings.HasPrefix(includePath, "/usr") ||
			strings.HasPrefix(includePath, "/opt") ||
			strings.HasPrefix(includePath, "/Library") ||
			strings.Contains(includePath, "sdk") {
			continue
		}

		// Check if it's a relative path (project-local)
		if !filepath.IsAbs(includePath) || strings.HasPrefix(includePath, ".") {
			// Collect all headers from this directory
			err := filepath.Walk(includePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip errors
				}
				if info.IsDir() {
					return nil
				}

				ext := filepath.Ext(path)
				if ext == ".h" || ext == ".hpp" || ext == ".hxx" || ext == ".hh" {
					content, err := os.ReadFile(path)
					if err != nil {
						return nil // Skip unreadable files
					}

					// Use relative path as key
					relPath, _ := filepath.Rel(includePath, path)
					if relPath == "" {
						relPath = filepath.Base(path)
					}
					includeFiles[relPath] = content
				}
				return nil
			})
			if err != nil {
				log.Debug().Err(err).Str("path", includePath).Msg("Error walking include path")
			}
		}
	}

	return includeFiles, nil
}

// generateCacheKeyRaw generates a cache key from raw source.
func (s *Service) generateCacheKeyRaw(req *Request, rawSource []byte) string {
	key := &cache.CompilationKey{
		Compiler:   req.Args.Compiler,
		TargetArch: req.TargetArch.String(),
		Flags:      req.Args.Flags,
		Defines:    req.Args.Defines,
		SourceHash: cache.HashBytes(rawSource),
	}
	return key.Build()
}

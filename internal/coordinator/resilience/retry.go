package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Common errors
var (
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	ErrNotRetryable       = errors.New("error is not retryable")
)

// RetryConfig holds retry configuration.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries uint64
	// InitialInterval is the initial retry interval.
	InitialInterval time.Duration
	// Multiplier is the backoff multiplier.
	Multiplier float64
	// MaxInterval is the maximum retry interval.
	MaxInterval time.Duration
	// MaxElapsedTime is the maximum total time for all retries.
	MaxElapsedTime time.Duration
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     5 * time.Second,
		MaxElapsedTime:  30 * time.Second,
	}
}

// RetryOperation represents an operation that can be retried.
type RetryOperation func() error

// RetryableOperation represents an operation that returns a result.
type RetryableOperation[T any] func() (T, error)

// Retry executes an operation with exponential backoff.
func Retry(ctx context.Context, cfg RetryConfig, operation RetryOperation) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.Multiplier = cfg.Multiplier
	b.MaxInterval = cfg.MaxInterval
	b.MaxElapsedTime = cfg.MaxElapsedTime

	// Wrap with max retries
	bWithRetries := backoff.WithMaxRetries(b, cfg.MaxRetries)

	// Wrap with context
	bWithContext := backoff.WithContext(bWithRetries, ctx)

	attempt := 0
	return backoff.Retry(func() error {
		attempt++
		err := operation()
		if err != nil {
			if !IsRetryable(err) {
				log.Debug().
					Int("attempt", attempt).
					Err(err).
					Msg("Non-retryable error, stopping retries")
				return backoff.Permanent(err)
			}
			log.Debug().
				Int("attempt", attempt).
				Err(err).
				Msg("Retryable error, will retry")
		}
		return err
	}, bWithContext)
}

// RetryWithResult executes an operation with exponential backoff and returns a result.
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, operation RetryableOperation[T]) (T, error) {
	var result T
	var lastErr error

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.Multiplier = cfg.Multiplier
	b.MaxInterval = cfg.MaxInterval
	b.MaxElapsedTime = cfg.MaxElapsedTime

	bWithRetries := backoff.WithMaxRetries(b, cfg.MaxRetries)
	bWithContext := backoff.WithContext(bWithRetries, ctx)

	attempt := 0
	err := backoff.Retry(func() error {
		attempt++
		var opErr error
		result, opErr = operation()
		if opErr != nil {
			lastErr = opErr
			if !IsRetryable(opErr) {
				log.Debug().
					Int("attempt", attempt).
					Err(opErr).
					Msg("Non-retryable error, stopping retries")
				return backoff.Permanent(opErr)
			}
			log.Debug().
				Int("attempt", attempt).
				Err(opErr).
				Msg("Retryable error, will retry")
			return opErr
		}
		return nil
	}, bWithContext)

	if err != nil {
		return result, err
	}
	return result, lastErr
}

// IsRetryable determines if an error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors (not retryable)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check gRPC status codes
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.OK:
			return false
		case codes.Canceled:
			return false
		case codes.InvalidArgument:
			return false
		case codes.NotFound:
			return false
		case codes.AlreadyExists:
			return false
		case codes.PermissionDenied:
			return false
		case codes.FailedPrecondition:
			return false
		case codes.Unimplemented:
			return false
		case codes.Unauthenticated:
			return false
		// These are retryable
		case codes.Unknown:
			return true
		case codes.DeadlineExceeded:
			return true
		case codes.ResourceExhausted:
			return true
		case codes.Aborted:
			return true
		case codes.Internal:
			return true
		case codes.Unavailable:
			return true
		case codes.DataLoss:
			return true
		}
	}

	// Default: retry on unknown errors
	return true
}

// RetryNotify is like Retry but calls a notify function on each retry.
func RetryNotify(ctx context.Context, cfg RetryConfig, operation RetryOperation, notify func(err error, duration time.Duration)) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.Multiplier = cfg.Multiplier
	b.MaxInterval = cfg.MaxInterval
	b.MaxElapsedTime = cfg.MaxElapsedTime

	bWithRetries := backoff.WithMaxRetries(b, cfg.MaxRetries)
	bWithContext := backoff.WithContext(bWithRetries, ctx)

	return backoff.RetryNotify(func() error {
		err := operation()
		if err != nil && !IsRetryable(err) {
			return backoff.Permanent(err)
		}
		return err
	}, bWithContext, notify)
}

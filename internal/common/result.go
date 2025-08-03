package common

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Result represents a generic result with success/failure information
type Result[T any] struct {
	Value    T
	Error    error
	Duration time.Duration
	Metadata map[string]interface{}
}

// IsSuccess returns true if the result represents a successful operation
func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

// Unwrap returns the value if successful, or panics with the error
func (r Result[T]) Unwrap() T {
	if r.Error != nil {
		panic(fmt.Sprintf("attempted to unwrap failed result: %v", r.Error))
	}
	return r.Value
}

// UnwrapOr returns the value if successful, or the provided default
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.Error != nil {
		return defaultValue
	}
	return r.Value
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger func(string, ...interface{})
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger func(string, ...interface{})) *ErrorHandler {
	if logger == nil {
		logger = log.Printf
	}
	return &ErrorHandler{logger: logger}
}

// Handle processes an error with context and returns a formatted error
func (h *ErrorHandler) Handle(operation string, err error, context ...interface{}) error {
	if err == nil {
		return nil
	}

	formattedErr := fmt.Errorf("%s failed: %w", operation, err)

	if len(context) > 0 {
		h.logger("Error in %s with context %+v: %v", operation, context, err)
	} else {
		h.logger("Error in %s: %v", operation, err)
	}

	return formattedErr
}

// ExecuteWithTiming executes a function and returns a timed result
func ExecuteWithTiming[T any](fn func() (T, error)) Result[T] {
	start := time.Now()
	value, err := fn()
	duration := time.Since(start)

	return Result[T]{
		Value:    value,
		Error:    err,
		Duration: duration,
		Metadata: make(map[string]interface{}),
	}
}

// ExecuteWithContext executes a function with context and timeout
func ExecuteWithContext[T any](ctx context.Context, timeout time.Duration, fn func(context.Context) (T, error)) Result[T] {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	type resultChan struct {
		value T
		err   error
	}

	resultCh := make(chan resultChan, 1)
	start := time.Now()

	go func() {
		value, err := fn(ctx)
		resultCh <- resultChan{value: value, err: err}
	}()

	select {
	case result := <-resultCh:
		return Result[T]{
			Value:    result.value,
			Error:    result.err,
			Duration: time.Since(start),
			Metadata: make(map[string]interface{}),
		}
	case <-ctx.Done():
		var zero T
		return Result[T]{
			Value:    zero,
			Error:    fmt.Errorf("operation timed out: %w", ctx.Err()),
			Duration: time.Since(start),
			Metadata: make(map[string]interface{}),
		}
	}
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	BackoffBase time.Duration
	BackoffMax  time.Duration
	ShouldRetry func(error) bool
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BackoffBase: 100 * time.Millisecond,
		BackoffMax:  5 * time.Second,
		ShouldRetry: func(err error) bool {
			// Default: retry on any error
			return err != nil
		},
	}
}

// ExecuteWithRetry executes a function with retry logic
func ExecuteWithRetry[T any](config RetryConfig, fn func() (T, error)) Result[T] {
	var lastErr error
	start := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		value, err := fn()
		if err == nil {
			return Result[T]{
				Value:    value,
				Error:    nil,
				Duration: time.Since(start),
				Metadata: map[string]interface{}{
					"attempts": attempt,
				},
			}
		}

		lastErr = err

		// Don't retry if we shouldn't or if this was the last attempt
		if !config.ShouldRetry(err) || attempt == config.MaxAttempts {
			break
		}

		// Calculate backoff with exponential increase
		backoff := config.BackoffBase * time.Duration(1<<uint(attempt-1))
		if backoff > config.BackoffMax {
			backoff = config.BackoffMax
		}

		time.Sleep(backoff)
	}

	var zero T
	return Result[T]{
		Value:    zero,
		Error:    fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr),
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"attempts": config.MaxAttempts,
			"failed":   true,
		},
	}
}

// CollectMapValues collects all values from the map into a slice with proper synchronization
func CollectMapValues[K comparable, V any](mu *sync.RWMutex, m map[K]V) []V {
	mu.RLock()
	defer mu.RUnlock()
	
	values := make([]V, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	
	return values
}

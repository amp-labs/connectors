package future

import (
	"context"
	"runtime/debug"

	"github.com/amp-labs/connectors/common/utils"
)

// Executor is responsible for executing asynchronous operations and resolving promises.
// It abstracts the goroutine creation and panic recovery logic used by Future/Promise.
type Executor[T any] interface {
	// Go executes the callback in a new goroutine and resolves the promise with the result.
	// Panics are recovered and converted to errors.
	Go(promise *Promise[T], callback func() (T, error))

	// GoContext executes the callback with a context in a new goroutine and resolves the promise.
	// The context allows for cancellation and timeout handling.
	GoContext(ctx context.Context, promise *Promise[T], callback func(ctx context.Context) (T, error))
}

// DefaultGoExecutor is the default implementation of Executor that spawns goroutines
// for async execution with panic recovery.
type DefaultGoExecutor[T any] struct{}

// Compile-time check to ensure DefaultGoExecutor implements Executor.
var _ Executor[any] = (*DefaultGoExecutor[any])(nil)

// Go executes the callback asynchronously in a new goroutine.
// Any panics in the callback are recovered and converted to errors.
func (e *DefaultGoExecutor[T]) Go(promise *Promise[T], callback func() (T, error)) {
	go func() {
		// Panic recovery MUST be deferred to catch panics in fn()
		defer func() {
			if err := recover(); err != nil {
				// Convert panic to error with stack trace for debugging
				pre := utils.GetPanicRecoveryError(err, debug.Stack())

				promise.Failure(pre)
			}
		}()

		// Execute the user's function and complete the promise
		value, err := callback()
		promise.Complete(value, err)
	}()
}

// GoContext executes the callback asynchronously with context support.
// Creates a child context that is canceled when the goroutine completes to prevent leaks.
// Any panics in the callback are recovered and converted to errors.
//
//nolint:contextcheck // GoContext intentionally creates a new cancellable context for the goroutine
func (e *DefaultGoExecutor[T]) GoContext(
	ctx context.Context, promise *Promise[T], callback func(ctx context.Context) (T, error),
) {
	// Ensure we always have a valid context
	if ctx == nil {
		ctx = context.Background()
	}

	// Create a child context that we can cancel when the goroutine completes
	// This prevents context leaks and allows independent cancellation
	goCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				// Convert panic to error with stack trace
				pre := utils.GetPanicRecoveryError(err, debug.Stack())

				promise.Failure(pre)
			}

			// CRITICAL: Always cancel the context to prevent leaks
			// This releases resources even if the function panicked
			cancel()
		}()

		// Execute user's function with the child context
		value, err := callback(goCtx)

		promise.Complete(value, err)
	}()
}

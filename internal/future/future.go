// Package future provides a Future/Promise implementation for asynchronous programming in Go.
//
// This package follows the split responsibility pattern where Future is the read-only side
// and Promise is the write-only side of an asynchronous computation. This design prevents
// consumers from accidentally completing a future they should only be reading from.
//
// Key design principles:
//   - Futures are immutable after completion (write-once semantics)
//   - All operations are goroutine-safe and can be called from multiple goroutines
//   - Results are memoized - once computed, they're stored and reused
//   - Panic recovery is built-in to prevent goroutine crashes
//   - Context support for cancellation and timeouts
//
// Example usage:
//
//	// Using Go() for simple async operations
//	future := future.Go(func() (int, error) {
//	    // expensive computation
//	    return 42, nil
//	})
//	result, err := future.Await()
//
//	// Using New() for manual control
//	fut, promise := future.New[string]()
//	go func() {
//	    result := someAsyncWork()
//	    promise.Success(result)
//	}()
//	value, err := fut.Await()
package future

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/atomic"

	"github.com/amp-labs/connectors/common/contexts"
	"github.com/amp-labs/connectors/common/try"
)

var (
	// ErrNilFuture is returned when a nil future is provided to a function.
	ErrNilFuture = errors.New("nil future provided")
	// ErrNilFunction is returned when a nil function is provided to a function.
	ErrNilFunction = errors.New("nil function provided")
)

// Future represents the read-only side of an asynchronous computation.
// It provides methods to await the result of a computation that may not yet be complete.
//
// Design notes:
//   - The future can only be completed once (enforced by sync.Once)
//   - All read operations (Await, AwaitContext) are idempotent and goroutine-safe
//   - The result is memoized after first completion
//   - Uses a closed channel (resultReady) as a broadcast mechanism for completion
//   - Multiple goroutines can safely await the same future
//
// Thread safety:
//   - Concurrent calls to Await/AwaitContext are safe
//   - The sync.Once ensures the result is only set once
//   - The channel pattern allows multiple readers to unblock simultaneously
type Future[T any] struct {
	once        sync.Once     // Ensures result is only written once
	result      try.Try[T]    // Stores the final result (value + error)
	resultReady chan struct{} // Closed when result is available (broadcast signal)
	cancelFunc  func()        // A callback to cancel the underlying computation, if supported
}

// New creates a new Future/Promise pair for manual async computation management.
//
// This function returns both the read-side (Future) and write-side (Promise) of the
// computation. The caller is responsible for:
//   - Managing goroutine lifecycle (New doesn't launch goroutines)
//   - Ensuring the promise is eventually fulfilled (via Success, Failure, or Complete)
//   - Handling any concurrency concerns in their code
//
// Use this when you need fine-grained control over when and how the computation runs.
// For simpler cases, use Go() or GoContext() instead.
//
// Example:
//
//	fut, promise := future.New[int]()
//	go func() {
//	    time.Sleep(time.Second)
//	    promise.Success(42)
//	}()
//	result, err := fut.Await()
//
// Design note: The unbuffered channel is intentionally not closed here - it will be
// closed by the Promise when fulfill() is called, which broadcasts to all waiters.
func New[T any](cancel ...func()) (future *Future[T], promise *Promise[T]) {
	future = &Future[T]{
		resultReady: make(chan struct{}), // Unbuffered - closed on completion
	}

	promise = &Promise[T]{
		future:      future,
		canceled:    atomic.NewBool(false),
		cancelFuncs: cancel,
	}

	future.cancelFunc = promise.cancel

	return future, promise
}

// Go creates a new Future and executes the given function in a new goroutine.
//
// This is the most common way to create a future. It's a convenience wrapper that:
//   - Creates a Future/Promise pair
//   - Launches a goroutine to execute the function
//   - Automatically fulfills the promise with the result
//   - Catches any panics and converts them to errors
//
// The panic recovery includes stack traces for debugging, making it safer than raw
// goroutines while still being convenient.
//
// Example:
//
//	future := future.Go(func() (string, error) {
//	    result, err := http.Get("https://api.example.com")
//	    if err != nil {
//	        return "", err
//	    }
//	    return result.Body, nil
//	})
//	data, err := future.Await()
//
// Design note: The panic recovery is critical because panics in goroutines crash the
// entire program by default. This converts them to errors that can be handled normally.
func Go[T any](fn func() (T, error)) *Future[T] {
	return GoWithExecutor[T](&DefaultGoExecutor[T]{}, fn)
}

// GoWithExecutor creates a new Future and executes the function using a custom executor.
//
// This allows you to customize how the async operation is executed, such as for testing
// or using a different concurrency model. Most users should use Go() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[int]{}
//	future := future.GoWithExecutor(customExec, func() (int, error) {
//	    return 42, nil
//	})
func GoWithExecutor[T any](exec Executor[T], fn func() (T, error)) *Future[T] {
	future, promise := New[T]()

	exec.Go(promise, fn)

	return future
}

// GoContext creates a new Future and executes the given function in a goroutine with context support.
//
// This is the context-aware version of Go(). It:
//   - Creates a cancellable child context for the goroutine
//   - Passes that context to the user's function
//   - Automatically cancels the context when the goroutine completes
//   - Catches panics and converts them to errors
//
// The child context allows the async operation to be canceled independently and ensures
// proper cleanup when the computation finishes.
//
// Example:
//
//	ctx := context.Background()
//	future := future.GoContext(ctx, func(ctx context.Context) (Data, error) {
//	    return fetchDataWithContext(ctx)
//	})
//	result, err := future.AwaitContext(ctx)
//
// Design notes:
//   - A nil context is automatically replaced with context.Background()
//   - The child context (goCtx) is canceled in defer to prevent context leaks
//   - The cancel is called even on panic to ensure cleanup
//   - The parent context can still cancel the child via the usual context mechanisms
//
//nolint:contextcheck // GoContext intentionally creates a new cancellable context for the goroutine
func GoContext[T any](ctx context.Context, operation func(context.Context) (T, error)) *Future[T] {
	return GoContextWithExecutor[T](ctx, &DefaultGoExecutor[T]{}, operation)
}

// GoContextWithExecutor creates a new Future with context support using a custom executor.
//
// This combines the context-awareness of GoContext with the flexibility of custom executors.
// Most users should use GoContext() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[int]{}
//	future := future.GoContextWithExecutor(ctx, customExec, func(ctx context.Context) (int, error) {
//	    return fetchWithContext(ctx)
//	})
//
//nolint:contextcheck // GoContextWithExecutor intentionally creates a new cancellable context for the goroutine
func GoContextWithExecutor[T any](
	ctx context.Context, exec Executor[T], operation func(context.Context) (T, error),
) *Future[T] {
	// Ensure we always have a valid context
	if ctx == nil {
		ctx = context.Background()
	}

	// Create a child context that we can cancel when the goroutine completes
	// This prevents context leaks and allows independent cancellation
	goCtx, cancel := context.WithCancel(ctx)

	// Create the Future/Promise pair
	future, promise := New[T](cancel)

	exec.GoContext(goCtx, promise, operation)

	return future
}

// Await blocks until the future completes and returns the result.
//
// This is the primary way to retrieve the result of an async computation.
//
// Behavior:
//   - Blocks the calling goroutine until the future completes
//   - If already complete, returns immediately with the memoized result
//   - Can be called multiple times - always returns the same result
//   - Safe for concurrent use by multiple goroutines
//
// The blocking is implemented via channel receive, which is efficient and allows
// the Go scheduler to do other work while waiting.
//
// Example:
//
//	future := future.Go(expensiveComputation)
//	result, err := future.Await()  // Blocks until complete
//	result2, err2 := future.Await() // Returns immediately with same result
func (f *Future[T]) Await() (T, error) { //nolint:ireturn
	// Wait for completion by receiving from the closed channel
	// This unblocks when the channel is closed in Promise.fulfill()
	<-f.resultReady

	// Return the memoized result - this is always the same value
	return f.result.Get()
}

// AwaitContext blocks until the future completes or the context is canceled.
//
// This is the context-aware version of Await, allowing for timeouts and cancellation.
//
// Behavior:
//   - Returns the result if the future completes first
//   - Returns context.Canceled/DeadlineExceeded if context is canceled first
//   - If future is already complete, returns result immediately (ignoring context state)
//   - If ctx is nil, behaves like Await()
//   - Safe for concurrent use by multiple goroutines
//
// IMPORTANT: This does NOT cancel the underlying computation - it only stops waiting.
// The future will continue computing in the background. If you need to cancel the
// computation itself, use GoContext and cancel the context you passed to it.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	future := future.Go(slowComputation)
//	result, err := future.AwaitContext(ctx)
//	if errors.Is(err, context.DeadlineExceeded) {
//	    // Timeout occurred
//	}
//
// Design note: The select statement races the context cancellation against future
// completion. Whichever happens first wins. The channel close is instant, so there's
// no meaningful race condition to worry about.
func (f *Future[T]) AwaitContext(ctx context.Context) (T, error) { //nolint:ireturn
	// Nil context means no cancellation - just use regular Await
	if ctx == nil {
		return f.Await()
	}

	// Fast path: check if already completed to avoid blocking on context
	select {
	case <-f.resultReady:
		// Future already completed - return result immediately
		return f.result.Get()
	default:
	}

	// Race the context cancellation against future completion
	select {
	case <-f.resultReady:
		// Future completed before context cancellation
		return f.result.Get()
	case <-ctx.Done():
		// Context was canceled/timed out before completion
		var zero T

		return zero, ctx.Err()
	}
}

// ToChannel returns a channel that will receive the result when the future completes.
//
// This allows integration with select statements and channel-based workflows.
func (f *Future[T]) ToChannel() <-chan try.Try[T] {
	ch := make(chan try.Try[T], 1)

	go func() {
		val, err := f.Await()
		ch <- try.Try[T]{Value: val, Error: err}

		close(ch)
	}()

	return ch
}

// ToChannelContext is the context-aware version of Channel.
//
// This allows waiting for the future result with context support, enabling
// cancellation and timeouts when using the channel in select statements.
func (f *Future[T]) ToChannelContext(ctx context.Context) <-chan try.Try[T] {
	ch := make(chan try.Try[T], 1)

	go func() {
		val, err := f.AwaitContext(ctx)
		ch <- try.Try[T]{Value: val, Error: err}
		close(ch)
	}()

	return ch
}

// Cancel attempts to cancel the underlying computation if it supports cancellation.
//
// This is a best-effort operation. If the future was created with GoContext,
// this will cancel that context. If the future was created  with Go or New
// without a cancellable context, this does nothing.
//
// This simply sends a signal. The actual cancellation depends on whether the
// underlying operation respects context cancellation. If it doesn't, the operation
// will continue running in the background.
func (f *Future[T]) Cancel() {
	if f.cancelFunc != nil {
		f.cancelFunc()
	}
}

// NewError creates a Future that is already completed with the given error.
//
// This is a convenience function for creating pre-failed futures, which is useful for:
//   - Early returns in error cases
//   - Validation failures that should be async-compatible
//   - Propagating errors through future-based APIs
//
// The returned future is immediately complete, so Await() will return instantly.
//
// Example:
//
//	func fetchUser(id string) *Future[User] {
//	    if id == "" {
//	        return NewError[User](errors.New("id cannot be empty"))
//	    }
//	    return Go(func() (User, error) {
//	        return db.GetUser(id)
//	    })
//	}
func NewError[T any](err error) *Future[T] {
	future, promise := New[T]()

	// Immediately complete the future with an error
	promise.Failure(err)

	return future
}

// Map transforms a Future[A] into a Future[B] by applying a function to the successful result.
//
// This is a functional programming primitive for chaining async operations. It allows
// you to transform the result of a future without manually handling the error cases.
//
// Behavior:
//   - If the original future succeeds, applies fn to the value and returns the result
//   - If the original future fails, propagates the error without calling fn
//   - If fn returns an error, that error is propagated
//   - Launches a new goroutine via Go() to perform the transformation
//
// Example:
//
//	// Convert user ID to user object
//	idFuture := future.Go(getUserId)
//	userFuture := future.Map(idFuture, func(id int) (User, error) {
//	    return fetchUser(id)
//	})
//	user, err := userFuture.Await()
//
// Design notes:
//   - Returns a pre-failed future if inputs are invalid (nil checks)
//   - Uses Go() internally, so transformation happens in a separate goroutine
//   - Error propagation is automatic - no need for manual if err != nil checks
func Map[A, B any](fut *Future[A], transformFunc func(A) (B, error)) *Future[B] {
	return MapWithExecutor[A, B](fut, &DefaultGoExecutor[B]{}, transformFunc)
}

// MapWithExecutor transforms a Future[A] into a Future[B] using a custom executor.
//
// This is identical to Map but allows you to specify a custom executor for the
// transformation. Most users should use Map() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[User]{}
//	idFuture := future.Go(getUserId)
//	userFuture := future.MapWithExecutor(idFuture, customExec, func(id int) (User, error) {
//	    return fetchUser(id)
//	})
func MapWithExecutor[A, B any](
	fut *Future[A],
	exec Executor[B],
	transformFunc func(A) (B, error),
) *Future[B] {
	// Input validation - return pre-failed futures for invalid inputs
	if fut == nil {
		return NewError[B](ErrNilFuture)
	}

	if transformFunc == nil {
		return NewError[B](ErrNilFunction)
	}

	// Create a new future that awaits the original and transforms the result
	return GoWithExecutor[B](exec, func() (B, error) {
		// Wait for the original future to complete
		val, err := fut.Await()
		if err != nil {
			// Propagate the error without calling fn
			var zero B

			return zero, err
		}

		// Apply the transformation function to the successful value
		return transformFunc(val)
	})
}

// MapContext is the context-aware version of Map.
//
// This is identical to Map but with context support, allowing:
//   - Cancellation of the transformation via context
//   - Passing context to the transformation function (e.g., for DB calls)
//   - Timeout support for the entire operation
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	idFuture := future.Go(getUserId)
//	userFuture := future.MapContext(ctx, idFuture, func(ctx context.Context, id int) (User, error) {
//	    return db.FetchUserWithContext(ctx, id)
//	})
//	user, err := userFuture.AwaitContext(ctx)
//
// Design note: Uses GoContext internally to respect the context throughout the operation.
func MapContext[A, B any](
	ctx context.Context, fut *Future[A], transformFunc func(context.Context, A) (B, error),
) *Future[B] {
	return MapContextWithExecutor[A, B](ctx, fut, &DefaultGoExecutor[B]{}, transformFunc)
}

// MapContextWithExecutor transforms a Future[A] into a Future[B] with context support using a custom executor.
//
// This combines context-awareness with custom executor support for transformations.
// Most users should use MapContext() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[User]{}
//	idFuture := future.Go(getUserId)
//	userFuture := future.MapContextWithExecutor(ctx, idFuture, customExec,
//	    func(ctx context.Context, id int) (User, error) {
//	        return fetchUserWithContext(ctx, id)
//	    })
func MapContextWithExecutor[A, B any](
	ctx context.Context,
	fut *Future[A],
	exec Executor[B],
	transformFunc func(context.Context, A) (B, error),
) *Future[B] {
	// Input validation
	if fut == nil {
		return NewError[B](ErrNilFuture)
	}

	if transformFunc == nil {
		return NewError[B](ErrNilFunction)
	}

	// Create a context-aware future that awaits and transforms
	return GoContextWithExecutor[B](ctx, exec, func(ctx context.Context) (B, error) {
		// Wait for the original future with context support
		val, err := fut.AwaitContext(ctx)
		if err != nil {
			// Propagate errors (including context cancellation)
			var zero B

			return zero, err
		}

		// Apply transformation with context
		return transformFunc(ctx, val)
	})
}

// FlatMap transforms a Future[A] into a Future[B] by applying a function that returns a Future[B].
//
// This is the "monadic bind" operation for futures. It's used when the transformation
// itself is asynchronous. The key difference from Map is that fn returns a *Future[B],
// not a plain B, preventing nested futures (Future[Future[B]]).
//
// Use cases:
//   - Chaining multiple async operations sequentially
//   - Dependent async calls where the second call needs the result of the first
//   - Building complex async workflows
//
// Example:
//
//	// Fetch user, then fetch their posts (two async operations)
//	userFuture := future.Go(fetchUser)
//	postsFuture := future.FlatMap(userFuture, func(user User) *Future[[]Post] {
//	    return future.Go(func() ([]Post, error) {
//	        return fetchPosts(user.ID)
//	    })
//	})
//	posts, err := postsFuture.Await()
//
// Design note: Without FlatMap, you'd get Future[Future[[]Post]] which is awkward.
// FlatMap "flattens" this to just Future[[]Post] by awaiting both futures.
func FlatMap[A, B any](fut *Future[A], fn func(A) *Future[B]) *Future[B] {
	return FlatMapWithExecutor[A, B](fut, &DefaultGoExecutor[B]{}, fn)
}

// FlatMapWithExecutor transforms a Future[A] into a Future[B] using a custom executor.
//
// This is identical to FlatMap but allows you to specify a custom executor.
// Most users should use FlatMap() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]Post]{}
//	userFuture := future.Go(fetchUser)
//	postsFuture := future.FlatMapWithExecutor(userFuture, customExec, func(user User) *Future[[]Post] {
//	    return future.Go(func() ([]Post, error) {
//	        return fetchPosts(user.ID)
//	    })
//	})
func FlatMapWithExecutor[A, B any](fut *Future[A], exec Executor[B], transform func(A) *Future[B]) *Future[B] {
	return GoWithExecutor(exec, func() (B, error) {
		// Await the first future
		val, err := fut.Await()
		if err != nil {
			var zero B

			return zero, err
		}

		// Apply transform to get the second future, then await it
		// This "flattens" Future[Future[B]] to Future[B]
		return transform(val).Await()
	})
}

// FlatMapContext is the context-aware version of FlatMap.
//
// Identical to FlatMap but with context support throughout the chain.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	userFuture := future.GoContext(ctx, fetchUser)
//	postsFuture := future.FlatMapContext(ctx, userFuture, func(user User) *Future[[]Post] {
//	    return future.GoContext(ctx, func(ctx context.Context) ([]Post, error) {
//	        return fetchPostsWithContext(ctx, user.ID)
//	    })
//	})
//	posts, err := postsFuture.AwaitContext(ctx)
//
// Design note: The same context is used for awaiting both futures, allowing
// cancellation at any point in the chain.
func FlatMapContext[A, B any](ctx context.Context, fut *Future[A], fn func(A) *Future[B]) *Future[B] {
	return FlatMapContextWithExecutor[A, B](ctx, fut, &DefaultGoExecutor[B]{}, fn)
}

// FlatMapContextWithExecutor transforms a Future[A] into a Future[B] with context support using a custom executor.
//
// This combines context-awareness with custom executor support for chaining async operations.
// Most users should use FlatMapContext() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]Post]{}
//	userFuture := future.GoContext(ctx, fetchUser)
//	postsFuture := future.FlatMapContextWithExecutor(ctx, userFuture, customExec, func(user User) *Future[[]Post] {
//	    return future.GoContext(ctx, func(ctx context.Context) ([]Post, error) {
//	        return fetchPostsWithContext(ctx, user.ID)
//	    })
//	})
func FlatMapContextWithExecutor[A, B any](
	ctx context.Context, fut *Future[A], exec Executor[B], transform func(A) *Future[B],
) *Future[B] {
	return GoContextWithExecutor[B](ctx, exec, func(ctx context.Context) (B, error) {
		// Await first future with context
		val, err := fut.AwaitContext(ctx)
		if err != nil {
			var zero B

			return zero, err
		}

		// Await second future with context (flattening)
		return transform(val).AwaitContext(ctx)
	})
}

// Combine combines multiple futures into a single future that completes when all inputs complete.
//
// This is useful for waiting on multiple independent async operations that can run in parallel.
//
// Behavior:
//   - Waits for ALL futures to complete
//   - Returns results in the same order as input futures
//   - Short-circuits on first error (doesn't wait for remaining futures)
//   - Returns empty slice if no futures provided
//
// IMPORTANT: The futures can be running concurrently - this just waits for them
// sequentially. To get true parallelism, launch the futures BEFORE calling Combine.
//
// Example:
//
//	// Launch three concurrent operations
//	fut1 := future.Go(fetchUser)
//	fut2 := future.Go(fetchPosts)
//	fut3 := future.Go(fetchComments)
//
//	// Wait for all to complete
//	combined := future.Combine(fut1, fut2, fut3)
//	results, err := combined.Await()
//	if err != nil {
//	    // One of the futures failed
//	}
//	user, posts, comments := results[0], results[1], results[2]
//
// Design notes:
//   - Short-circuiting on error is intentional for fail-fast behavior
//   - The futures themselves keep running in background even after error
//   - For non-short-circuiting behavior, use CombineNoShortCircuit
//   - The goroutine waits sequentially but futures run concurrently
func Combine[T any](futures ...*Future[T]) *Future[[]T] {
	return CombineWithExecutor[T](&DefaultGoExecutor[[]T]{}, futures...)
}

// CombineWithExecutor combines multiple futures using a custom executor.
//
// This is identical to Combine but allows you to specify a custom executor for the
// combination operation. Most users should use Combine() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]User]{}
//	fut1 := future.Go(fetchUser1)
//	fut2 := future.Go(fetchUser2)
//	combined := future.CombineWithExecutor(customExec, fut1, fut2)
func CombineWithExecutor[T any](exec Executor[[]T], futures ...*Future[T]) *Future[[]T] {
	future, promise := New[[]T]()

	// Special case: no futures to combine
	if len(futures) == 0 {
		promise.Success(nil)

		return future
	}

	exec.Go(promise, func() ([]T, error) {
		// Pre-allocate slice for efficiency
		results := make([]T, 0, len(futures))

		// Await each future in order
		for _, fut := range futures {
			val, err := fut.Await()
			if err != nil {
				// Short-circuit: fail immediately on first error
				return nil, err
			}

			results = append(results, val)
		}

		// All futures succeeded
		return results, nil
	})

	return future
}

// CombineContext is the context-aware version of Combine.
//
// Identical to Combine but with context support, allowing cancellation while waiting.
//
// Behavior:
//   - Checks context before awaiting each future
//   - Short-circuits on context cancellation OR first error
//   - Returns context.Canceled or context.DeadlineExceeded if canceled
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	fut1 := future.GoContext(ctx, fetchUser)
//	fut2 := future.GoContext(ctx, fetchPosts)
//	fut3 := future.GoContext(ctx, fetchComments)
//
//	combined := future.CombineContext(ctx, fut1, fut2, fut3)
//	results, err := combined.AwaitContext(ctx)
//	if errors.Is(err, context.DeadlineExceeded) {
//	    // Timeout occurred while waiting
//	}
//
// Design note: The context check happens between awaits to allow early exit if
// the context is canceled while waiting for slow futures.
func CombineContext[T any](ctx context.Context, futures ...*Future[T]) *Future[[]T] {
	return CombineContextWithExecutor[T](ctx, &DefaultGoExecutor[[]T]{}, futures...)
}

// CombineContextWithExecutor combines multiple futures with context support using a custom executor.
//
// This is identical to CombineContext but allows you to specify a custom executor.
// Most users should use CombineContext() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]User]{}
//	fut1 := future.GoContext(ctx, fetchUser1)
//	fut2 := future.GoContext(ctx, fetchUser2)
//	combined := future.CombineContextWithExecutor(ctx, customExec, fut1, fut2)
func CombineContextWithExecutor[T any](
	ctx context.Context, exec Executor[[]T], futures ...*Future[T],
) *Future[[]T] {
	future, promise := New[[]T]()

	if len(futures) == 0 {
		promise.Success(nil)

		return future
	}

	exec.GoContext(ctx, promise, func(ctx context.Context) ([]T, error) {
		results := make([]T, 0, len(futures))

		for _, fut := range futures {
			// Check if context was canceled before awaiting next future
			if !contexts.IsContextAlive(ctx) {
				return nil, ctx.Err()
			}

			// Await with context support
			val, err := fut.AwaitContext(ctx)
			if err != nil {
				// Fail on first error (could be context error or future error)
				return nil, err
			}

			results = append(results, val)
		}

		return results, nil
	})

	return future
}

// CombineNoShortCircuit combines multiple futures, collecting ALL results and errors.
//
// Unlike Combine, this waits for ALL futures to complete even if some fail. This is useful
// when you need to know about all failures, or when you want partial results.
//
// Behavior:
//   - Waits for ALL futures to complete (never short-circuits)
//   - Collects ALL errors and aggregates them with errors.Join
//   - Returns ALL results (including zero values for failed futures)
//   - If any errors occurred, the combined future fails with joined errors
//
// IMPORTANT: The error case returns BOTH the results AND the error. To access the
// results when there are errors, you'd need to use the try.Try type directly via
// the Promise.fulfill mechanism. However, when using Await(), you only get the error.
//
// Example:
//
//	// Try to fetch multiple users - we want to know which ones failed
//	futs := make([]*Future[User], len(ids))
//	for i, id := range ids {
//	    futs[i] = future.Go(func() (User, error) { return fetchUser(id) })
//	}
//
//	combined := future.CombineNoShortCircuit(futs...)
//	results, err := combined.Await()
//	if err != nil {
//	    // Multiple errors may have been joined
//	    log.Printf("Some fetches failed: %v", err)
//	    // Note: results is nil when using Await() with errors
//	}
//
// Design note: This uses promise.fulfill with a Try type to store both results and
// errors, but Await() only returns the error part in failure cases.
func CombineNoShortCircuit[T any](futures ...*Future[T]) *Future[[]T] {
	return CombineNoShortCircuitWithExecutor[T](&DefaultGoExecutor[[]T]{}, futures...)
}

// CombineNoShortCircuitWithExecutor combines multiple futures without short-circuiting using a custom executor.
//
// This is identical to CombineNoShortCircuit but allows you to specify a custom executor.
// Most users should use CombineNoShortCircuit() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]User]{}
//	futs := []*Future[User]{fut1, fut2, fut3}
//	combined := future.CombineNoShortCircuitWithExecutor(customExec, futs...)
func CombineNoShortCircuitWithExecutor[T any](exec Executor[[]T], futures ...*Future[T]) *Future[[]T] {
	future, promise := New[[]T]()

	if len(futures) == 0 {
		promise.Success([]T{})

		return future
	}

	exec.Go(promise, func() ([]T, error) {
		// Collect both results and errors
		results := make([]T, 0, len(futures))
		errs := make([]error, 0, len(futures))

		// Await ALL futures regardless of errors
		for _, fut := range futures {
			val, err := fut.Await()
			if err != nil {
				errs = append(errs, err)
			}

			// Always collect results (may include zero values)
			results = append(results, val)
		}

		// If there were any errors, join them all together
		if len(errs) > 0 {
			err := errors.Join(errs...)

			// Use internal fulfill to store both results AND error
			// (Await() will only return the error, but Promise has both)
			promise.fulfill(try.Try[[]T]{
				Value: results,
				Error: err,
			})

			return results, err
		}

		return results, nil
	})

	return future
}

// CombineContextNoShortCircuit is the context-aware version of CombineNoShortCircuit.
//
// Waits for ALL futures with context support, collecting all results and errors.
//
// Behavior:
//   - Checks context before each future await
//   - If context is canceled, immediately returns context error (does NOT wait for remaining futures)
//   - If context stays alive, waits for ALL futures and joins all errors
//
// NOTE: Unlike CombineNoShortCircuit, this DOES short-circuit on context cancellation
// (but not on future errors).
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	futs := make([]*Future[Result], len(tasks))
//	for i, task := range tasks {
//	    futs[i] = future.GoContext(ctx, task)
//	}
//
//	combined := future.CombineContextNoShortCircuit(ctx, futs...)
//	results, err := combined.AwaitContext(ctx)
//	if err != nil {
//	    // Could be context error OR joined future errors
//	}
//
// Design note: Context cancellation causes immediate return (short-circuit), but
// future errors do not - all futures are awaited to collect all errors.
func CombineContextNoShortCircuit[T any](ctx context.Context, futures ...*Future[T]) *Future[[]T] {
	return CombineContextNoShortCircuitWithExecutor[T](ctx, &DefaultGoExecutor[[]T]{}, futures...)
}

// CombineContextNoShortCircuitWithExecutor combines futures with context support without
// short-circuiting using a custom executor.
//
// This is identical to CombineContextNoShortCircuit but allows you to specify a custom executor.
// Most users should use CombineContextNoShortCircuit() instead.
//
// Example:
//
//	customExec := &MyCustomExecutor[[]Result]{}
//	futs := []*Future[Result]{fut1, fut2, fut3}
//	combined := future.CombineContextNoShortCircuitWithExecutor(ctx, customExec, futs...)
func CombineContextNoShortCircuitWithExecutor[T any](
	ctx context.Context, exec Executor[[]T], futures ...*Future[T],
) *Future[[]T] {
	future, promise := New[[]T]()

	if len(futures) == 0 {
		promise.Success([]T{})

		return future
	}

	exec.GoContext(ctx, promise, func(ctx context.Context) ([]T, error) {
		results := make([]T, 0, len(futures))
		errs := make([]error, 0, len(futures))

		for _, fut := range futures {
			// Short-circuit on context cancellation
			if !contexts.IsContextAlive(ctx) {
				return nil, ctx.Err()
			}

			// Await with context (may fail due to context or future error)
			val, err := fut.AwaitContext(ctx)
			if err != nil {
				// Collect the error but keep going
				errs = append(errs, err)
			}

			results = append(results, val)
		}

		// Join all errors if any occurred
		if len(errs) > 0 {
			err := errors.Join(errs...)

			promise.fulfill(try.Try[[]T]{
				Value: results,
				Error: err,
			})

			return results, err
		}

		return results, nil
	})

	return future
}

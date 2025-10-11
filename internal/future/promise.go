package future

import (
	"github.com/amp-labs/connectors/common/try"
	"go.uber.org/atomic"
)

// Promise represents the write-only side of an asynchronous computation.
//
// A Promise is used to complete a Future by providing either a successful value
// or an error. It's the "producer" side while Future is the "consumer" side.
//
// Key guarantees:
//   - A promise can only be fulfilled once (enforced by sync.Once in the future)
//   - Multiple calls to Success/Failure/Complete are safe (later calls are ignored)
//   - Fulfillment is thread-safe and can happen from any goroutine
//   - Fulfilling a promise unblocks all goroutines waiting on the associated future
//
// Design note: The promise holds a reference to the future, not the other way around.
// This ensures futures can be passed around without exposing the ability to complete them.
type Promise[T any] struct {
	future      *Future[T]   // Reference to the associated future (write access)
	canceled    *atomic.Bool // Atomic flag indicating if this promise has been canceled
	cancelFuncs []func()     // Functions to call when promise is canceled
}

// IsCancelled returns true if the promise has been canceled.
//
// This is a thread-safe check that can be called from any goroutine.
// Once a promise is canceled, it remains canceled permanently.
func (p *Promise[T]) IsCancelled() bool {
	return p.canceled.Load()
}

// cancel marks the promise as canceled and executes all registered cancel functions.
//
// This is an internal method used by the future package for cancellation propagation.
// It uses atomic compare-and-swap to ensure cancel functions are only executed once,
// even if cancel() is called multiple times concurrently.
//
// Thread safety: Safe to call from any goroutine. Multiple calls are idempotent -
// only the first call executes the cancel functions.
func (p *Promise[T]) cancel() {
	// Only execute cancel functions once using atomic CAS
	if p.canceled.CompareAndSwap(false, true) {
		for _, cancel := range p.cancelFuncs {
			cancel()
		}
	}
}

// fulfill is the internal method that actually completes the promise.
//
// This is the core mechanism for promise completion. It:
//   - Stores the result (value + error) in the future
//   - Closes the resultReady channel to broadcast completion
//   - Is idempotent (safe to call multiple times)
//
// Thread safety is provided by sync.Once, which ensures only the first call succeeds.
// The defer+recover protects against double-close panics, though sync.Once should
// prevent them anyway.
//
// Design notes:
//   - Uses try.Try[T] to store both value and error together
//   - Channel close is a broadcast - all waiters are unblocked simultaneously
//   - Recover is defensive programming - shouldn't be needed but provides safety
//   - This is internal (unexported) - callers use Success/Failure/Complete
func (p *Promise[T]) fulfill(result try.Try[T]) {
	defer func() {
		// Defensive: recover from any panic (e.g., double close)
		// This shouldn't happen due to sync.Once, but provides safety
		_ = recover()
	}()

	// Only the first call to once.Do executes - others are no-ops
	p.future.once.Do(func() {
		// Store the result for later retrieval
		p.future.result = result

		// Close channel to broadcast completion to all waiters
		// A closed channel immediately returns to all receivers
		close(p.future.resultReady)
	})
}

// Success fulfills the promise with a successful value.
//
// Use this when the async computation succeeded and you have a value to provide.
//
// Example:
//
//	fut, promise := future.New[string]()
//	go func() {
//	    result := doWork()
//	    promise.Success(result)
//	}()
//
// Thread safety: Safe to call from any goroutine. If called multiple times,
// only the first call takes effect.
func (p *Promise[T]) Success(value T) {
	p.fulfill(try.Try[T]{
		Value: value,
		Error: nil,
	})
}

// Failure fulfills the promise with an error.
//
// Use this when the async computation failed and you need to propagate the error.
//
// Example:
//
//	fut, promise := future.New[User]()
//	go func() {
//	    user, err := fetchUser(id)
//	    if err != nil {
//	        promise.Failure(err)
//	        return
//	    }
//	    promise.Success(user)
//	}()
//
// Design note: The value is set to the zero value of T. This is necessary because
// the try.Try[T] type requires both a value and error, but only the error matters
// in the failure case.
//
// Thread safety: Safe to call from any goroutine. If called multiple times,
// only the first call takes effect.
func (p *Promise[T]) Failure(err error) {
	var zero T // Zero value of T (e.g., 0 for int, "" for string, nil for pointers)

	p.fulfill(try.Try[T]{
		Value: zero,
		Error: err,
	})
}

// Complete fulfills the promise with a value and error pair.
//
// This is a convenience method that matches Go's standard (value, error) return pattern.
// It internally calls either Success or Failure based on the error.
//
// Use this when you have both a value and error from a function call, following
// Go's idiomatic error handling.
//
// Example:
//
//	fut, promise := future.New[Data]()
//	go func() {
//	    // Function returns (Data, error) tuple
//	    data, err := fetchData()
//	    // Complete handles both cases
//	    promise.Complete(data, err)
//	}()
//
// Behavior:
//   - If err != nil: calls Failure(err), ignoring the value
//   - If err == nil: calls Success(value)
//
// Design note: This is the most commonly used method because it matches Go's
// error handling conventions. It's what Go() uses internally.
//
// Thread safety: Safe to call from any goroutine. If called multiple times,
// only the first call takes effect.
func (p *Promise[T]) Complete(value T, err error) {
	if err != nil {
		p.Failure(err)
	} else {
		p.Success(value)
	}
}

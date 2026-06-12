package simultaneously

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/amp-labs/connectors/common/contexts"
)

// Job is a function that performs a unit of work and returns an error if it fails.
type Job func(ctx context.Context) error

// ErrPanicRecovered is the base error for panic recovery.
var ErrPanicRecovered = errors.New("panic recovered")

// Do runs the given functions in parallel and returns the first error encountered.
// See SimultaneouslyCtx for more information.
func Do(maxConcurrent int, f ...Job) error {
	return DoCtx(context.Background(), maxConcurrent, f...)
}

// DoCtx runs the given functions in parallel and returns the first error encountered.
// If no error is encountered, it returns nil. In the event that an error happens, all other functions
// are canceled (via their context) to hopefully save on CPU cycles. It's up to the individual functions
// to check their context and return early if they are canceled.
//
// The maxConcurrent parameter is used to limit the number of functions that run at the same time.
// If maxConcurrent is less than 1, all functions will run at the same time.
//
// Panics that occur within the callback functions are automatically recovered and converted to errors.
// This prevents a single panicking function from crashing the entire process.
func DoCtx(ctx context.Context, maxConcurrent int, callback ...Job) error {
	ctx, cancel := context.WithCancel(ctx)

	var cancelOnce sync.Once
	defer cancelOnce.Do(cancel)

	if maxConcurrent < 1 {
		maxConcurrent = len(callback)
	}

	exec := newExecutor(maxConcurrent, &cancelOnce, cancel)
	defer exec.cleanup()

	exec.launchAll(ctx, callback)

	errs := exec.collectResults(len(callback))

	return combineErrors(errs)
}

// executor manages the concurrent execution of callback functions.
type executor struct {
	cancelOnce *sync.Once
	cancel     context.CancelFunc
	sem        chan struct{}
	errorChan  chan error
	doneChan   chan struct{}
	waitGroup  sync.WaitGroup
}

// newExecutor creates a new executor with the given concurrency limit.
func newExecutor(maxConcurrent int, cancelOnce *sync.Once, cancel context.CancelFunc) *executor {
	sem := make(chan struct{}, maxConcurrent)

	// Fill the semaphore with maxConcurrent empty structs
	for range maxConcurrent {
		sem <- struct{}{}
	}

	return &executor{
		cancelOnce: cancelOnce,
		cancel:     cancel,
		sem:        sem,
		errorChan:  make(chan error, maxConcurrent),
		doneChan:   make(chan struct{}, maxConcurrent),
	}
}

// cleanup closes all channels after waiting for goroutines to finish.
func (e *executor) cleanup() {
	e.waitGroup.Wait()
	close(e.sem)
	close(e.errorChan)
	close(e.doneChan)
}

// launchAll starts all callback functions in separate goroutines.
func (e *executor) launchAll(ctx context.Context, callbacks []Job) {
	for _, fn := range callbacks {
		e.waitGroup.Add(1)
		go e.run(ctx, fn)
	}
}

// run executes a single callback function with semaphore control and panic recovery.
func (e *executor) run(ctx context.Context, fn Job) {
	<-e.sem // take one out (will block if empty)

	defer func() {
		e.sem <- struct{}{} // put it back
		e.waitGroup.Done()
	}()

	defer e.recoverPanic()

	e.executeCallback(ctx, fn)
}

// recoverPanic recovers from panics and converts them to errors.
func (e *executor) recoverPanic() {
	r := recover()
	if r == nil {
		return
	}

	err := formatPanicError(r)

	// Cancel the context to stop other functions
	e.cancelOnce.Do(e.cancel)

	// Send the panic as an error
	e.errorChan <- err
}

// formatPanicError converts a panic value into a formatted error with stack trace.
func formatPanicError(r any) error {
	if e, ok := r.(error); ok {
		return fmt.Errorf("%w: %w\n%s", ErrPanicRecovered, e, debug.Stack())
	}

	return fmt.Errorf("%w: %v\n%s", ErrPanicRecovered, r, debug.Stack())
}

// executeCallback runs the callback function and sends the result to the appropriate channel.
func (e *executor) executeCallback(ctx context.Context, fn func(context.Context) error) {
	if !contexts.IsContextAlive(ctx) {
		e.errorChan <- ctx.Err()

		return
	}

	err := fn(ctx)
	if err != nil {
		e.cancelOnce.Do(e.cancel)
		e.errorChan <- err
	} else {
		e.doneChan <- struct{}{}
	}
}

// collectResults waits for all goroutines to complete and collects errors.
func (e *executor) collectResults(count int) []error {
	var errs []error

	for range count {
		select {
		case err := <-e.errorChan:
			errs = append(errs, err)
		case <-e.doneChan: // Function completed successfully
		}
	}

	return errs
}

// combineErrors returns a single error from a slice of errors.
func combineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.Join(errs...)
	}
}

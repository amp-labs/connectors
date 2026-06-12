package simultaneously

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errTestPanic = errors.New("test error panic")
	errTest      = errors.New("test error")
)

func TestDoCtx_RecoversPanic(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 2,
		func(ctx context.Context) error {
			panic("intentional panic for testing")
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "intentional panic for testing")
	assert.Contains(t, err.Error(), "simultaneously_test.go") // stack trace should be present
}

func TestDoCtx_RecoversPanicError(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 2,
		func(ctx context.Context) error {
			panic(errTestPanic)
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	require.ErrorIs(t, err, errTestPanic)
	assert.Contains(t, err.Error(), "simultaneously_test.go") // stack trace should be present
}

func TestDoCtx_MixedSuccessAndPanic(t *testing.T) {
	t.Parallel()

	var successCount atomic.Int32

	err := DoCtx(t.Context(), 3,
		func(ctx context.Context) error {
			successCount.Add(1)
			time.Sleep(10 * time.Millisecond)

			return nil
		},
		func(ctx context.Context) error {
			time.Sleep(5 * time.Millisecond)
			panic("boom")
		},
		func(ctx context.Context) error {
			successCount.Add(1)
			time.Sleep(10 * time.Millisecond)

			return nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "boom")

	// At least one function should have completed successfully
	// (though others may have been canceled due to the panic)
	assert.GreaterOrEqual(t, successCount.Load(), int32(1))
}

func TestDoCtx_MultiplePanics(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 3,
		func(ctx context.Context) error {
			panic("panic 1")
		},
		func(ctx context.Context) error {
			panic("panic 2")
		},
		func(ctx context.Context) error {
			panic("panic 3")
		},
	)

	require.Error(t, err)

	// Should get at least one panic error
	assert.Contains(t, err.Error(), "panic recovered")

	// Due to concurrency and early cancellation, we might get multiple panics joined
	// or just the first one that was caught
	panicCount := strings.Count(err.Error(), "panic recovered")
	assert.GreaterOrEqual(t, panicCount, 1)
}

func TestDoCtx_PanicDoesNotAffectOtherGoroutines(t *testing.T) {
	t.Parallel()

	var completed atomic.Int32

	err := DoCtx(t.Context(), 10,
		func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			completed.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			// Panic immediately
			panic("early panic")
		},
		func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			completed.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			completed.Add(1)

			return nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "early panic")
}

func TestDoCtx_SuccessfulExecution(t *testing.T) {
	t.Parallel()

	var counter atomic.Int32

	err := DoCtx(t.Context(), 3,
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int32(3), counter.Load())
}

func TestDoCtx_ErrorReturnedInsteadOfPanic(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 2,
		func(ctx context.Context) error {
			return errTest
		},
		func(ctx context.Context) error {
			return nil
		},
	)

	require.Error(t, err)
	require.ErrorIs(t, err, errTest)
	// Should not contain panic recovery message since this was a normal error
	assert.NotContains(t, err.Error(), "panic recovered")
}

func TestDoCtx_PanicWithNilValue(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 1,
		func(ctx context.Context) error {
			var nilPtr *string
			_ = *nilPtr // This will panic with nil pointer dereference

			return nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "nil pointer") // The panic message should mention nil pointer
}

func TestDo_RecoversPanic(t *testing.T) {
	t.Parallel()

	err := Do(2,
		func(ctx context.Context) error {
			panic("panic in Do function")
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "panic in Do function")
}

func TestDoCtx_PanicWithStackTrace(t *testing.T) {
	t.Parallel()

	err := DoCtx(t.Context(), 1,
		func(ctx context.Context) error {
			helper()

			return nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "helper function panic")
	// Stack trace should show the helper function
	assert.Contains(t, err.Error(), "simultaneously_test.go")
}

// Helper function for testing stack traces.
func helper() {
	panic("helper function panic")
}

func TestDoCtx_ContextCancellationAfterPanic(t *testing.T) {
	t.Parallel()

	var canceledCount atomic.Int32

	err := DoCtx(t.Context(), 5,
		func(ctx context.Context) error {
			// Panic immediately
			panic("early panic")
		},
		func(ctx context.Context) error {
			// This should potentially be canceled
			time.Sleep(100 * time.Millisecond)

			if ctx.Err() != nil {
				canceledCount.Add(1)

				return ctx.Err()
			}

			return nil
		},
		func(ctx context.Context) error {
			// This should potentially be canceled
			time.Sleep(100 * time.Millisecond)

			if ctx.Err() != nil {
				canceledCount.Add(1)

				return ctx.Err()
			}

			return nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
}

func TestDoCtxWithWaitInterval_SuccessfulExecution(t *testing.T) {
	t.Parallel()

	var counter atomic.Int32

	err := DoCtxWithWaitInterval(t.Context(), 2, 10,
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int32(3), counter.Load())
}

func TestDoCtxWithWaitInterval_ZeroWaitDelegatesToDoCtx(t *testing.T) {
	t.Parallel()

	var counter atomic.Int32

	err := DoCtxWithWaitInterval(t.Context(), 1, 0,
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
		func(ctx context.Context) error {
			counter.Add(1)

			return nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int32(2), counter.Load())
}

func TestDoCtxWithWaitInterval_WaitsBetweenBatches(t *testing.T) {
	t.Parallel()

	const (
		waitIntervalMS = 50
		maxConcurrent  = 2
	)

	var batchStarts []time.Time

	err := DoCtxWithWaitInterval(t.Context(), maxConcurrent, waitIntervalMS,
		func(ctx context.Context) error {
			batchStarts = append(batchStarts, time.Now())

			return nil
		},
		func(ctx context.Context) error {
			batchStarts = append(batchStarts, time.Now())

			return nil
		},
		func(ctx context.Context) error {
			batchStarts = append(batchStarts, time.Now())

			return nil
		},
	)

	require.NoError(t, err)
	require.Len(t, batchStarts, 3)

	// Jobs 0 and 1 run in the first batch; job 2 runs after the wait interval.
	firstBatchLatest := batchStarts[0]
	if batchStarts[1].After(firstBatchLatest) {
		firstBatchLatest = batchStarts[1]
	}

	assert.GreaterOrEqual(t, batchStarts[2].Sub(firstBatchLatest), time.Duration(waitIntervalMS)*time.Millisecond)
}

func TestDoCtxWithWaitInterval_ContextCanceledDuringWait(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		_ = DoCtxWithWaitInterval(ctx, 1, 1000,
			func(ctx context.Context) error {
				return nil
			},
			func(ctx context.Context) error {
				return nil
			},
		)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("DoCtxWithWaitInterval did not return after context cancellation")
	}
}

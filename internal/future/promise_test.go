package future

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromise_Success(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Success(42)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPromise_Failure(t *testing.T) {
	t.Parallel()

	fut, promise := New[string]()

	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errTest, err)
	assert.Equal(t, "", result)
}

func TestPromise_Complete_Success(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Complete(42, nil)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPromise_Complete_Failure(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Complete(0, errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errTest, err)
	assert.Equal(t, 0, result)
}

func TestPromise_Complete_IgnoresValueOnError(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	// Even though we pass 42, the error should take precedence
	promise.Complete(42, errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errTest, err)
	assert.Equal(t, 0, result) // Should be zero value, not 42
}

func TestPromise_MultipleSuccess_FirstWins(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Success(1)
	promise.Success(2)
	promise.Success(3)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 1, result) // First call wins
}

func TestPromise_MultipleFailure_FirstWins(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Failure(errTest)
	promise.Failure(errOriginal)
	promise.Failure(errTransform)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errTest, err) // First call wins
	assert.Equal(t, 0, result)
}

func TestPromise_MixedComplete_FirstWins(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Success(42)
	promise.Failure(errTest)
	promise.Complete(99, nil)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result) // First call (Success) wins
}

func TestPromise_ConcurrentSuccess(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Launch multiple goroutines trying to complete the promise
	for i := range numGoroutines {
		go func(val int) {
			defer waitGroup.Done()
			promise.Success(val)
		}(i)
	}

	waitGroup.Wait()

	// One of the values should win
	result, err := fut.Await()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, result, 0)
	assert.Less(t, result, numGoroutines)
}

func TestPromise_ConcurrentFailure(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Launch multiple goroutines trying to complete the promise with errors
	for range numGoroutines {
		go func() {
			defer waitGroup.Done()
			promise.Failure(errTest)
		}()
	}

	waitGroup.Wait()

	// Should get the error
	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errTest, err)
	assert.Equal(t, 0, result)
}

func TestPromise_ConcurrentMixed(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Mix of success and failure calls
	for i := range numGoroutines {
		go func(val int) {
			defer waitGroup.Done()

			if val%2 == 0 {
				promise.Success(val)
			} else {
				promise.Failure(errTest)
			}
		}(i)
	}

	waitGroup.Wait()

	// Should complete without panic
	result, err := fut.Await()

	// Either success or failure, but not both
	if err == nil {
		assert.GreaterOrEqual(t, result, 0)
		assert.Less(t, result, numGoroutines)
	} else {
		assert.Equal(t, errTest, err)
		assert.Equal(t, 0, result)
	}
}

func TestPromise_IsCancelled_Initial(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	assert.False(t, promise.IsCancelled())
}

func TestPromise_IsCancelled_AfterCancel(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	promise.cancel()

	assert.True(t, promise.IsCancelled())
}

func TestPromise_Cancel_Idempotent(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	// Cancel multiple times should be safe
	promise.cancel()
	promise.cancel()
	promise.cancel()

	assert.True(t, promise.IsCancelled())
}

func TestPromise_Cancel_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Multiple goroutines canceling concurrently
	for range numGoroutines {
		go func() {
			defer waitGroup.Done()
			promise.cancel()
		}()
	}

	waitGroup.Wait()

	assert.True(t, promise.IsCancelled())
}

func TestPromise_Cancel_CallsCancelFuncs(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	var callCount atomic.Int32

	// Add cancel functions
	promise.cancelFuncs = append(promise.cancelFuncs,
		func() { callCount.Add(1) },
		func() { callCount.Add(10) },
		func() { callCount.Add(100) },
	)

	promise.cancel()

	// All cancel functions should be called
	assert.Equal(t, int32(111), callCount.Load())
	assert.True(t, promise.IsCancelled())

	// Verify future is not affected by cancel (cancel doesn't complete the promise)
	select {
	case <-fut.resultReady:
		t.Fatal("future should not be completed by cancel")
	default:
	}
}

func TestPromise_Cancel_CallsFuncsOnce(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	var callCount atomic.Int32

	promise.cancelFuncs = append(promise.cancelFuncs,
		func() { callCount.Add(1) },
	)

	// Cancel multiple times
	promise.cancel()
	promise.cancel()
	promise.cancel()

	// Cancel functions should only be called once
	assert.Equal(t, int32(1), callCount.Load())
}

func TestPromise_Cancel_ConcurrentFuncExecution(t *testing.T) {
	t.Parallel()

	_, promise := New[int]()

	var callCount atomic.Int32

	promise.cancelFuncs = append(promise.cancelFuncs,
		func() { callCount.Add(1) },
	)

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Multiple goroutines trying to cancel
	for range numGoroutines {
		go func() {
			defer waitGroup.Done()
			promise.cancel()
		}()
	}

	waitGroup.Wait()

	// Function should only be called once despite concurrent cancels
	assert.Equal(t, int32(1), callCount.Load())
}

func TestPromise_Cancel_DoesNotCompleteFuture(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.cancel()

	// Future should not be completed
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	_, err := fut.AwaitContext(ctx)

	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestPromise_Cancel_ThenComplete(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.cancel()

	assert.True(t, promise.IsCancelled())

	// Should still be able to complete the promise
	promise.Success(42)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPromise_Complete_ThenCancel(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	promise.Success(42)

	var callCount atomic.Int32

	promise.cancelFuncs = append(promise.cancelFuncs,
		func() { callCount.Add(1) },
	)

	promise.cancel()

	// Cancel function should still execute
	assert.Equal(t, int32(1), callCount.Load())
	assert.True(t, promise.IsCancelled())

	// Future should have the success value
	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPromise_FulfillPanicRecovery(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	// First fulfillment
	promise.Success(42)

	// This should not panic even though it tries to fulfill again
	assert.NotPanics(t, func() {
		promise.Success(99)
	})

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPromise_ZeroValueTypes(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		testPromiseZeroValueInt(t)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		testPromiseZeroValueString(t)
	})

	t.Run("pointer", func(t *testing.T) {
		t.Parallel()

		testPromiseZeroValuePointer(t)
	})

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		testPromiseZeroValueSlice(t)
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		testPromiseZeroValueStruct(t)
	})
}

func testPromiseZeroValueInt(t *testing.T) {
	t.Helper()

	fut, promise := New[int]()
	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, 0, result)
}

func testPromiseZeroValueString(t *testing.T) {
	t.Helper()

	fut, promise := New[string]()
	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, "", result)
}

func testPromiseZeroValuePointer(t *testing.T) {
	t.Helper()

	fut, promise := New[*int]()
	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Nil(t, result)
}

func testPromiseZeroValueSlice(t *testing.T) {
	t.Helper()

	fut, promise := New[[]int]()
	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Nil(t, result)
}

func testPromiseZeroValueStruct(t *testing.T) {
	t.Helper()

	type MyStruct struct {
		A int
		B string
	}

	fut, promise := New[MyStruct]()
	promise.Failure(errTest)

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, MyStruct{}, result)
}

func TestPromise_SuccessWithZeroValue(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	// Should be able to successfully complete with zero value
	promise.Success(0)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 0, result)
}

func TestPromise_SuccessWithNil(t *testing.T) {
	t.Parallel()

	fut, promise := New[*int]()

	// Should be able to successfully complete with nil pointer
	promise.Success(nil)

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestPromise_MultipleWaiters(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	const numWaiters = 10
	results := make(chan int, numWaiters)
	errors := make(chan error, numWaiters)

	// Launch multiple waiters
	for range numWaiters {
		go func() {
			val, err := fut.Await()
			results <- val
			errors <- err
		}()
	}

	// Give waiters time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Complete the promise
	promise.Success(42)

	// All waiters should receive the same result
	for range numWaiters {
		result := <-results
		err := <-errors

		require.NoError(t, err)
		assert.Equal(t, 42, result)
	}
}

func TestPromise_CompleteInDifferentGoroutine(t *testing.T) {
	t.Parallel()

	fut, promise := New[string]()

	go func() {
		time.Sleep(10 * time.Millisecond)
		promise.Success("async result")
	}()

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, "async result", result)
}

func TestPromise_ImmediateCompletion(t *testing.T) {
	t.Parallel()

	fut, promise := New[int]()

	// Complete before waiting
	promise.Success(42)

	// Await should return immediately
	start := time.Now()

	result, err := fut.Await()

	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately")
}

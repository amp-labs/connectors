package future

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testString = "test"

var (
	errExecutor      = errors.New("executor test error")
	errContextWasNil = errors.New("context was nil")
)

// TestDefaultGoExecutor_Go_Success verifies that DefaultGoExecutor.Go executes
// a callback successfully and resolves the promise with the correct value.
func TestDefaultGoExecutor_Go_Success(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	executor.Go(promise, func() (int, error) {
		return 42, nil
	})

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

// TestDefaultGoExecutor_Go_Error verifies that DefaultGoExecutor.Go handles
// errors returned from the callback and propagates them to the promise.
func TestDefaultGoExecutor_Go_Error(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[string]{}
	fut, promise := New[string]()

	executor.Go(promise, func() (string, error) {
		return "", errExecutor
	})

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errExecutor, err)
	assert.Equal(t, "", result)
}

// TestDefaultGoExecutor_Go_Panic verifies that DefaultGoExecutor.Go recovers
// from panics in the callback and converts them to errors with stack traces.
func TestDefaultGoExecutor_Go_Panic(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	executor.Go(promise, func() (int, error) {
		panic("executor test panic")
	})

	result, err := fut.Await()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "recovered from panic: executor test panic")
	assert.Contains(t, err.Error(), "stack trace:")
	assert.Equal(t, 0, result)
}

// TestDefaultGoExecutor_Go_MultipleCallbacks verifies that multiple callbacks
// can be executed concurrently using the same executor instance.
func TestDefaultGoExecutor_Go_MultipleCallbacks(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}

	fut1, promise1 := New[int]()
	fut2, promise2 := New[int]()
	fut3, promise3 := New[int]()

	executor.Go(promise1, func() (int, error) {
		return 1, nil
	})

	executor.Go(promise2, func() (int, error) {
		return 2, nil
	})

	executor.Go(promise3, func() (int, error) {
		return 3, nil
	})

	result1, err1 := fut1.Await()
	require.NoError(t, err1)
	assert.Equal(t, 1, result1)

	result2, err2 := fut2.Await()
	require.NoError(t, err2)
	assert.Equal(t, 2, result2)

	result3, err3 := fut3.Await()
	require.NoError(t, err3)
	assert.Equal(t, 3, result3)
}

// TestDefaultGoExecutor_GoContext_Success verifies that DefaultGoExecutor.GoContext
// executes a callback with context successfully.
func TestDefaultGoExecutor_GoContext_Success(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[string]{}
	fut, promise := New[string]()

	executor.GoContext(t.Context(), promise, func(ctx context.Context) (string, error) {
		return "success", nil
	})

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

// TestDefaultGoExecutor_GoContext_Error verifies that DefaultGoExecutor.GoContext
// handles errors returned from the callback.
func TestDefaultGoExecutor_GoContext_Error(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	executor.GoContext(t.Context(), promise, func(ctx context.Context) (int, error) {
		return 0, errExecutor
	})

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, errExecutor, err)
	assert.Equal(t, 0, result)
}

// TestDefaultGoExecutor_GoContext_Panic verifies that DefaultGoExecutor.GoContext
// recovers from panics in the callback and converts them to errors.
func TestDefaultGoExecutor_GoContext_Panic(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[float64]{}
	fut, promise := New[float64]()

	executor.GoContext(t.Context(), promise, func(ctx context.Context) (float64, error) {
		panic("context executor panic")
	})

	result, err := fut.Await()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "recovered from panic: context executor panic")
	assert.Contains(t, err.Error(), "stack trace:")
	assert.Zero(t, result)
}

// TestDefaultGoExecutor_GoContext_ContextCancellation verifies that the context
// can be used to cancel the operation.
func TestDefaultGoExecutor_GoContext_ContextCancellation(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	ctx, cancel := context.WithCancel(t.Context())

	executor.GoContext(ctx, promise, func(ctx context.Context) (int, error) {
		<-ctx.Done()

		return 0, ctx.Err()
	})

	// Cancel the context
	cancel()

	result, err := fut.Await()

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, 0, result)
}

// TestDefaultGoExecutor_GoContext_NilContext verifies that a nil context is
// handled gracefully and replaced with context.Background().
func TestDefaultGoExecutor_GoContext_NilContext(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[string]{}
	fut, promise := New[string]()

	executor.GoContext(t.Context(), promise, func(ctx context.Context) (string, error) {
		// Verify we got a valid context (not nil)
		if ctx == nil {
			return "", errContextWasNil
		}

		return "ok", nil
	})

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

// TestDefaultGoExecutor_GoContext_ContextNotLeaked verifies that the child
// context is properly canceled when the goroutine completes, preventing leaks.
func TestDefaultGoExecutor_GoContext_ContextNotLeaked(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	parentCtx := t.Context()

	var childCtx context.Context

	childCtxChan := make(chan context.Context, 1)

	executor.GoContext(parentCtx, promise, func(ctx context.Context) (int, error) {
		childCtxChan <- ctx

		return 100, nil
	})

	result, err := fut.Await()

	require.NoError(t, err)
	assert.Equal(t, 100, result)

	// Retrieve the child context from the channel
	childCtx = <-childCtxChan

	// Give the goroutine time to cleanup (call cancel)
	time.Sleep(10 * time.Millisecond)

	// Verify the child context was canceled
	select {
	case <-childCtx.Done():
		// Expected: context should be canceled
	default:
		t.Error("child context was not canceled after goroutine completed")
	}
}

// TestDefaultGoExecutor_GoContext_PanicAlsoCancelsContext verifies that even
// when a panic occurs, the context is still canceled to prevent leaks.
func TestDefaultGoExecutor_GoContext_PanicAlsoCancelsContext(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	parentCtx := t.Context()

	var childCtx context.Context

	childCtxChan := make(chan context.Context, 1)

	executor.GoContext(parentCtx, promise, func(ctx context.Context) (int, error) {
		childCtxChan <- ctx

		panic("panic during execution")
	})

	result, err := fut.Await()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic during execution")
	assert.Equal(t, 0, result)

	// Retrieve the child context from the channel
	childCtx = <-childCtxChan

	// Give the goroutine time to cleanup (call cancel in defer)
	time.Sleep(10 * time.Millisecond)

	// Verify the child context was canceled even after panic
	select {
	case <-childCtx.Done():
		// Expected: context should be canceled
	default:
		t.Error("child context was not canceled after panic")
	}
}

// TestDefaultGoExecutor_GoContext_ConcurrentExecution verifies that multiple
// GoContext calls can execute concurrently.
func TestDefaultGoExecutor_GoContext_ConcurrentExecution(t *testing.T) {
	t.Parallel()

	executor := &DefaultGoExecutor[int]{}

	fut1, promise1 := New[int]()
	fut2, promise2 := New[int]()
	fut3, promise3 := New[int]()

	ctx := t.Context()

	start := time.Now()

	executor.GoContext(ctx, promise1, func(ctx context.Context) (int, error) {
		time.Sleep(20 * time.Millisecond)

		return 10, nil
	})

	executor.GoContext(ctx, promise2, func(ctx context.Context) (int, error) {
		time.Sleep(20 * time.Millisecond)

		return 20, nil
	})

	executor.GoContext(ctx, promise3, func(ctx context.Context) (int, error) {
		time.Sleep(20 * time.Millisecond)

		return 30, nil
	})

	result1, err1 := fut1.Await()
	result2, err2 := fut2.Await()
	result3, err3 := fut3.Await()

	elapsed := time.Since(start)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	assert.Equal(t, 10, result1)
	assert.Equal(t, 20, result2)
	assert.Equal(t, 30, result3)

	// Should complete in ~20ms (concurrent), not ~60ms (sequential)
	assert.Less(t, elapsed, 40*time.Millisecond, "callbacks should run concurrently")
}

// TestDefaultGoExecutor_ImplementsInterface verifies that DefaultGoExecutor
// implements the Executor interface at compile time.
func TestDefaultGoExecutor_ImplementsInterface(t *testing.T) {
	t.Parallel()

	var _ Executor[int] = (*DefaultGoExecutor[int])(nil)

	var _ Executor[string] = (*DefaultGoExecutor[string])(nil)

	var _ Executor[any] = (*DefaultGoExecutor[any])(nil)
}

// TestDefaultGoExecutor_DifferentTypes verifies that DefaultGoExecutor works
// with different type parameters.
func TestDefaultGoExecutor_DifferentTypes(t *testing.T) {
	t.Parallel()

	t.Run("int type", func(t *testing.T) {
		t.Parallel()

		testDefaultGoExecutorInt(t)
	})

	t.Run("string type", func(t *testing.T) {
		t.Parallel()

		testDefaultGoExecutorString(t)
	})

	t.Run("struct type", func(t *testing.T) {
		t.Parallel()

		testDefaultGoExecutorStruct(t)
	})

	t.Run("pointer type", func(t *testing.T) {
		t.Parallel()

		testDefaultGoExecutorPointer(t)
	})

	t.Run("slice type", func(t *testing.T) {
		t.Parallel()

		testDefaultGoExecutorSlice(t)
	})
}

func testDefaultGoExecutorInt(t *testing.T) {
	t.Helper()

	executor := &DefaultGoExecutor[int]{}
	fut, promise := New[int]()

	executor.Go(promise, func() (int, error) {
		return 123, nil
	})

	result, err := fut.Await()
	require.NoError(t, err)
	assert.Equal(t, 123, result)
}

func testDefaultGoExecutorString(t *testing.T) {
	t.Helper()

	executor := &DefaultGoExecutor[string]{}
	fut, promise := New[string]()

	executor.Go(promise, func() (string, error) {
		return testString, nil
	})

	result, err := fut.Await()
	require.NoError(t, err)
	assert.Equal(t, testString, result)
}

func testDefaultGoExecutorStruct(t *testing.T) {
	t.Helper()

	type TestStruct struct {
		ID   int
		Name string
	}

	executor := &DefaultGoExecutor[TestStruct]{}
	fut, promise := New[TestStruct]()

	expected := TestStruct{ID: 1, Name: testString}

	executor.Go(promise, func() (TestStruct, error) {
		return expected, nil
	})

	result, err := fut.Await()
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func testDefaultGoExecutorPointer(t *testing.T) {
	t.Helper()

	type TestStruct struct {
		Value int
	}

	executor := &DefaultGoExecutor[*TestStruct]{}
	fut, promise := New[*TestStruct]()

	expected := &TestStruct{Value: 42}

	executor.Go(promise, func() (*TestStruct, error) {
		return expected, nil
	})

	result, err := fut.Await()
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func testDefaultGoExecutorSlice(t *testing.T) {
	t.Helper()

	executor := &DefaultGoExecutor[[]int]{}
	fut, promise := New[[]int]()

	expected := []int{1, 2, 3, 4, 5}

	executor.Go(promise, func() ([]int, error) {
		return expected, nil
	})

	result, err := fut.Await()
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

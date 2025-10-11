package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPanicRecoveryError(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil panic value", func(t *testing.T) {
		t.Parallel()

		err := GetPanicRecoveryError(nil, nil)
		assert.NoError(t, err)
	})

	t.Run("wraps error panic value", func(t *testing.T) {
		t.Parallel()

		originalErr := errors.New("test error") //nolint:err113
		err := GetPanicRecoveryError(originalErr, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPanicRecovery)
		require.ErrorIs(t, err, originalErr)
		assert.Contains(t, err.Error(), "test error")
	})

	t.Run("formats string panic value", func(t *testing.T) {
		t.Parallel()

		err := GetPanicRecoveryError("panic message", nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPanicRecovery)
		assert.Contains(t, err.Error(), "panic message")
	})

	t.Run("includes stack trace when provided with error", func(t *testing.T) {
		t.Parallel()

		testGetPanicRecoveryErrorWithStack(t)
	})

	t.Run("includes stack trace when provided with string", func(t *testing.T) {
		t.Parallel()

		testGetPanicRecoveryErrorStringWithStack(t)
	})

	t.Run("handles integer panic value", func(t *testing.T) {
		t.Parallel()

		testGetPanicRecoveryErrorInteger(t)
	})

	t.Run("handles struct panic value", func(t *testing.T) {
		t.Parallel()

		testGetPanicRecoveryErrorStruct(t)
	})
}

func testGetPanicRecoveryErrorWithStack(t *testing.T) {
	t.Helper()

	originalErr := errors.New("test error") //nolint:err113
	stack := []byte("goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:10")
	err := GetPanicRecoveryError(originalErr, stack)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPanicRecovery)
	assert.Contains(t, err.Error(), "test error")
	assert.Contains(t, err.Error(), "stack trace:")
	assert.Contains(t, err.Error(), "goroutine 1")
}

func testGetPanicRecoveryErrorStringWithStack(t *testing.T) {
	t.Helper()

	stack := []byte("goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:10")
	err := GetPanicRecoveryError("panic message", stack)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPanicRecovery)
	assert.Contains(t, err.Error(), "panic message")
	assert.Contains(t, err.Error(), "stack trace:")
	assert.Contains(t, err.Error(), "goroutine 1")
}

func testGetPanicRecoveryErrorInteger(t *testing.T) {
	t.Helper()

	err := GetPanicRecoveryError(42, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPanicRecovery)
	assert.Contains(t, err.Error(), "42")
}

func testGetPanicRecoveryErrorStruct(t *testing.T) {
	t.Helper()

	type testStruct struct {
		Message string
	}

	err := GetPanicRecoveryError(testStruct{Message: "test"}, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPanicRecovery)
	assert.Contains(t, err.Error(), "test")
}

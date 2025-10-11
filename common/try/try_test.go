package try

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTry_IsSuccess(t *testing.T) {
	t.Parallel()

	t.Run("returns true when Error is nil", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 42, Error: nil}
		assert.True(t, tr.IsSuccess())
	})

	t.Run("returns false when Error is not nil", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 0, Error: errors.New("test error")} //nolint:err113
		assert.False(t, tr.IsSuccess())
	})
}

func TestTry_IsFailure(t *testing.T) {
	t.Parallel()

	t.Run("returns false when Error is nil", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 42, Error: nil}
		assert.False(t, tr.IsFailure())
	})

	t.Run("returns true when Error is not nil", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 0, Error: errors.New("test error")} //nolint:err113
		assert.True(t, tr.IsFailure())
	})
}

func TestTry_Get(t *testing.T) {
	t.Parallel()

	t.Run("returns value and nil error on success", func(t *testing.T) {
		t.Parallel()

		tr := Try[string]{Value: "hello", Error: nil}
		val, err := tr.Get()

		require.NoError(t, err)
		assert.Equal(t, "hello", val)
	})

	t.Run("returns zero value and error on failure", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("test error") //nolint:err113
		tr := Try[string]{Value: "hello", Error: expectedErr}
		val, err := tr.Get()

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", val)
	})

	t.Run("returns zero value for complex types on failure", func(t *testing.T) {
		t.Parallel()

		type customType struct {
			Field string
		}

		expectedErr := errors.New("test error") //nolint:err113
		tr := Try[customType]{Value: customType{Field: "test"}, Error: expectedErr}
		val, err := tr.Get()

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, customType{}, val)
	})
}

func TestTry_GetOrElse(t *testing.T) {
	t.Parallel()

	t.Run("returns value on success", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 42, Error: nil}
		result := tr.GetOrElse(100)

		assert.Equal(t, 42, result)
	})

	t.Run("returns default value on failure", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 42, Error: errors.New("test error")} //nolint:err113
		result := tr.GetOrElse(100)

		assert.Equal(t, 100, result)
	})

	t.Run("works with string type", func(t *testing.T) {
		t.Parallel()

		tr := Try[string]{Value: "", Error: errors.New("test error")} //nolint:err113
		result := tr.GetOrElse("default")

		assert.Equal(t, "default", result)
	})
}

func TestMap(t *testing.T) {
	t.Parallel()

	t.Run("transforms value on success", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 5, Error: nil}
		result := Map(tr, func(v int) (string, error) {
			return "Number: " + string(rune(v+'0')), nil
		})

		assert.True(t, result.IsSuccess())
		assert.Equal(t, "Number: 5", result.Value)
	})

	t.Run("propagates error from original Try", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("original error") //nolint:err113
		tr := Try[int]{Value: 5, Error: expectedErr}
		result := Map(tr, func(v int) (string, error) {
			return "should not be called", nil
		})

		assert.True(t, result.IsFailure())
		assert.Equal(t, expectedErr, result.Error)
	})

	t.Run("captures error from transform function", func(t *testing.T) {
		t.Parallel()

		tr := Try[int]{Value: 5, Error: nil}
		expectedErr := errors.New("transform error") //nolint:err113
		result := Map(tr, func(v int) (string, error) {
			return "", expectedErr
		})

		assert.True(t, result.IsFailure())
		assert.Equal(t, expectedErr, result.Error)
	})

	t.Run("transforms between different types", func(t *testing.T) {
		t.Parallel()

		tr := Try[string]{Value: "42", Error: nil}
		result := Map(tr, func(v string) (int, error) {
			return 42, nil
		})

		assert.True(t, result.IsSuccess())
		assert.Equal(t, 42, result.Value)
	})
}

func TestFlatMap(t *testing.T) {
	t.Parallel()

	t.Run("transforms value on success", func(t *testing.T) {
		t.Parallel()

		testFlatMapSuccess(t)
	})

	t.Run("propagates error from original Try", func(t *testing.T) {
		t.Parallel()

		testFlatMapPropagatesError(t)
	})

	t.Run("returns failure from transform function", func(t *testing.T) {
		t.Parallel()

		testFlatMapTransformError(t)
	})

	t.Run("chains multiple operations", func(t *testing.T) {
		t.Parallel()

		testFlatMapChaining(t)
	})

	t.Run("stops chain on first error", func(t *testing.T) {
		t.Parallel()

		testFlatMapStopsOnError(t)
	})
}

func testFlatMapSuccess(t *testing.T) {
	t.Helper()

	tr := Try[int]{Value: 5, Error: nil}
	result := FlatMap(tr, func(v int) Try[string] {
		return Try[string]{Value: "Number: 5", Error: nil}
	})

	assert.True(t, result.IsSuccess())
	assert.Equal(t, "Number: 5", result.Value)
}

func testFlatMapPropagatesError(t *testing.T) {
	t.Helper()

	expectedErr := errors.New("original error") //nolint:err113
	tr := Try[int]{Value: 5, Error: expectedErr}
	result := FlatMap(tr, func(v int) Try[string] {
		return Try[string]{Value: "should not be called", Error: nil}
	})

	assert.True(t, result.IsFailure())
	assert.Equal(t, expectedErr, result.Error)
}

func testFlatMapTransformError(t *testing.T) {
	t.Helper()

	tr := Try[int]{Value: 5, Error: nil}
	expectedErr := errors.New("transform error") //nolint:err113
	result := FlatMap(tr, func(v int) Try[string] {
		return Try[string]{Value: "", Error: expectedErr}
	})

	assert.True(t, result.IsFailure())
	assert.Equal(t, expectedErr, result.Error)
}

func testFlatMapChaining(t *testing.T) {
	t.Helper()

	tr := Try[int]{Value: 10, Error: nil}

	result := FlatMap(tr, func(v int) Try[int] {
		return Try[int]{Value: v * 2, Error: nil}
	})

	resultStr := FlatMap(result, func(v int) Try[string] {
		return Try[string]{Value: "Result: 20", Error: nil}
	})

	assert.True(t, resultStr.IsSuccess())
	assert.Equal(t, "Result: 20", resultStr.Value)
}

func testFlatMapStopsOnError(t *testing.T) {
	t.Helper()

	tr := Try[int]{Value: 10, Error: nil}
	expectedErr := errors.New("chain error") //nolint:err113

	result := FlatMap(tr, func(v int) Try[int] {
		return Try[int]{Value: 0, Error: expectedErr}
	})

	resultStr := FlatMap(result, func(v int) Try[string] {
		return Try[string]{Value: "should not be called", Error: nil}
	})

	assert.True(t, resultStr.IsFailure())
	assert.Equal(t, expectedErr, resultStr.Error)
}

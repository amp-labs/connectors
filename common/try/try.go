// Package try provides a Result type for error handling in Go.
// It encapsulates either a successful value or an error, inspired by functional programming patterns.
package try

// Try represents a computation that may either succeed with a value of type A or fail with an error.
// It provides a type-safe way to handle errors without relying on multiple return values.
type Try[A any] struct {
	Value A     // The successful result value (only valid when Error is nil)
	Error error // The error if the operation failed (nil indicates success)
}

// IsSuccess returns true if the Try contains a successful value (Error is nil).
func (t Try[A]) IsSuccess() bool {
	return t.Error == nil
}

// IsFailure returns true if the Try contains an error (Error is not nil).
func (t Try[A]) IsFailure() bool {
	return t.Error != nil
}

// Get returns the value and error as separate return values.
// If the Try is a failure, returns the zero value of type A and the error.
// If the Try is a success, returns the value and nil error.
func (t Try[A]) Get() (A, error) { //nolint:ireturn
	if t.IsFailure() {
		var zero A

		return zero, t.Error
	} else {
		return t.Value, nil
	}
}

// GetOrElse returns the value if successful, otherwise returns the provided default value.
// This is useful for providing fallback values when operations may fail.
func (t Try[A]) GetOrElse(defaultValue A) A { //nolint:ireturn
	if t.IsSuccess() {
		return t.Value
	} else {
		return defaultValue
	}
}

// Map transforms a Try[A] into a Try[B] by applying function f to the value if successful.
// If the original Try is a failure, the error is propagated to the result.
// If f returns an error, the result Try[B] will contain that error.
func Map[A, B any](t Try[A], f func(A) (B, error)) Try[B] {
	if t.IsSuccess() {
		val, err := f(t.Value)

		return Try[B]{Value: val, Error: err}
	} else {
		return Try[B]{Error: t.Error}
	}
}

// FlatMap transforms a Try[A] into a Try[B] by applying function f to the value if successful.
// Unlike Map, f returns a Try[B] directly, avoiding nested Try types (e.g., Try[Try[B]]).
// If the original Try is a failure, the error is propagated to the result.
// This enables chaining multiple operations that return Try types.
func FlatMap[A, B any](t Try[A], f func(A) Try[B]) Try[B] {
	if t.IsSuccess() {
		return f(t.Value)
	}

	return Try[B]{Error: t.Error}
}

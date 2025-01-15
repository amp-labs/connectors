package mockutils

import (
	"errors"
	"strings"
)

// ExpectedSubsetErrors represents a collection of errors expected to be found within a target wrapped error.
// Instead of constructing and comparing a full wrapped error stack, this `error type` focuses on verifying that
// critical errors are present in the wrapped Go error.
//
// By using this type, you dictate to the testing code to perform subset comparison via `errors.Is`.
type ExpectedSubsetErrors []error

// errorsAre returns true if each expected error is present within the target error object.
// It is similar to errors.Is() but is applied on list of expected errors.
func errorsAre(actualError error, expectedErrors ExpectedSubsetErrors) bool {
	for _, expected := range expectedErrors {
		if !errors.Is(actualError, expected) &&
			!strings.Contains(actualError.Error(), expected.Error()) {
			return false
		}
	}

	return true
}

func (e ExpectedSubsetErrors) Error() string {
	return errors.Join(e...).Error()
}

package testutils

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// StrError is a marker error type used in tests to indicate that
// an error should primarily be compared by its string content.
//
// Unlike ordinary errors that rely strictly on errors.Is matching,
// StrError allows fallback comparison via substring matching of the
// error message. This is useful when:
//
//   - the concrete error type is unstable or wrapped differently
//   - dynamic context is added to the error
//   - only the semantic message matters for the test
//
// Test-helpers interpret this type as a signal that textual comparison
// is acceptable when strict error identity does not match.
type StrError struct {
	text string
}

func (e StrError) Error() string {
	return e.text
}

// StringError creates a StrError from a plain string.
//
// Instead of asserting exact error identity, the testing utilities
// will allow message-based comparison when this marker type is used.
func StringError(text string) StrError {
	return StrError{text: text}
}

// CheckOutputWithError verifies both output values and returned errors
// produced by a test case.
//
// Special handling:
//   - If expectedErr is nil, any actual error fails the test.
//   - If expectedErr is a StrError, comparison allows either:
//   - errors.Is match, OR
//   - substring match of the error message.
//   - Otherwise, strict errors.Is comparison is required.
func CheckOutputWithError(t *testing.T, name string,
	expected any, expectedErr error,
	actual any, actualErr error,
) {
	t.Helper()

	CheckError(t, name, expectedErr, actualErr)
	CheckOutput(t, name, expected, actual)
}

func CheckError(t *testing.T, name string, expectedErr error, actualErr error) {
	t.Helper()

	// Ensure both sides agree on whether an error should exist.
	if expectedErr == nil {
		if actualErr != nil {
			t.Fatalf("%s: expected no error, got: (%v)", name, actualErr)
		}
	} else {
		if actualErr == nil {
			t.Fatalf("%s: expected error: (%v), got nil", name, expectedErr)
		}

		// Special marker handling: detect whether the expected error is a StrError marker.
		var strErr StrError
		if errors.As(expectedErr, &strErr) {
			// StrError allows flexible comparison:
			// prefer errors.Is, but fall back to message containment.
			if !errors.Is(actualErr, expectedErr) && !strings.Contains(actualErr.Error(), expectedErr.Error()) {
				t.Fatalf("%s: expected error: (%v), got: (%v)", name, expectedErr, actualErr)
			}
		} else {
			// Default behavior: strict semantic comparison using errors.Is.
			if !errors.Is(actualErr, expectedErr) {
				t.Fatalf("%s: expected error: (%v), got: (%v)", name, expectedErr, actualErr)
			}
		}
	}
}

func CheckOutput(t *testing.T, name string,
	expected any, actual any,
) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s: expected: (%v),\n got: (%v)", name, expected, actual)
	}
}

func CheckErrors(t *testing.T, name string, expectedErrs []error, actualErr error) {
	t.Helper()

	if actualErr != nil {
		if len(expectedErrs) == 0 {
			t.Fatalf("%s: expected no errors, got: (%v)", name, actualErr)
		}
	} else {
		// check that missing error is what is expected
		if len(expectedErrs) != 0 {
			t.Fatalf("%s: expected errors (%v), but got nothing", name, expectedErrs)
		}
	}

	// check every error
	for _, expectedErr := range expectedErrs {
		CheckError(t, name, expectedErr, actualErr)
	}
}

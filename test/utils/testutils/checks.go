package testutils

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func CheckOutputWithError(t *testing.T, name string,
	expected any, expectedErr error,
	actual any, actualErr error,
) {
	t.Helper()

	// check for actual error value
	if !errors.Is(actualErr, expectedErr) {
		t.Fatalf("%s: expected: (%v), got: (%v)", name, expectedErr, actualErr)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s: expected: (%v), got: (%v)", name, expected, actual)
	}
}

func CheckErrors(t *testing.T, name string, expectedErrs []error, actualErr error) {
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
		if !errors.Is(actualErr, expectedErr) && !strings.Contains(actualErr.Error(), expectedErr.Error()) {
			t.Fatalf("%s: expected Error: (%v), got: (%v)", name, expectedErr, actualErr)
		}
	}
}

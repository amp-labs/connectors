package testutils

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func CheckOutput(t *testing.T, name string,
	expected any, actual any,
) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s: expected: (%v),\n got: (%v)", name, expected, actual)
	}
}

func CheckOutputWithError(t *testing.T, name string,
	expected any, expectedErr error,
	actual any, actualErr error,
) {
	t.Helper()

	// check for actual error value
	if !errors.Is(actualErr, expectedErr) {
		t.Fatalf("%s: expected: (%v), got: (%v)", name, expectedErr, actualErr)
	}

	CheckOutput(t, name, expected, actual)
}

// CheckErrorAny validates that at least one expectedErr is present inside actual error.
func CheckErrorAny(t *testing.T, name string, expectedErrs []error, actualErr error) {
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

	var found bool
	for _, expectedErr := range expectedErrs {
		if errors.Is(actualErr, expectedErr) || strings.Contains(actualErr.Error(), expectedErr.Error()) {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("%s: expected one of: (%v), got: (%v)", name, expectedErrs, actualErr)
	}
}

// CheckError validates that expectedErr is present inside actual error.
func CheckError(t *testing.T, name string, expectedErr error, actualErr error) {
	if actualErr != nil {
		if expectedErr == nil {
			t.Fatalf("%s: expected no errors, got: (%v)", name, actualErr)
		}
	} else {
		// check that missing error is what is expected
		if expectedErr != nil {
			t.Fatalf("%s: expected error (%v), but got nothing", name, expectedErr)
		}
	}

	if !errors.Is(actualErr, expectedErr) && !strings.Contains(actualErr.Error(), expectedErr.Error()) {
		t.Fatalf("%s: expected Error: (%v), got: (%v)", name, expectedErr, actualErr)
	}
}

// CheckErrors validates that each expected error is present within actual error.
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

package testutils

import (
	"errors"
	"reflect"
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

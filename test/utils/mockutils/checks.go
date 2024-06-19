package mockutils

import (
	"fmt"
	"testing"
)

func DoesObjectCorrespondToString(object any, correspondent string) bool {
	if object == nil && len(correspondent) == 0 {
		return true
	}

	return fmt.Sprintf("%v", object) == correspondent
}

func NoErrors(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("failed to start test, %v", err)
	}
}

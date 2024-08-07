package mockutils

import (
	"fmt"
	"strconv"
	"testing"
)

func DoesObjectCorrespondToString(object any, correspondent string) bool {
	if object == nil && len(correspondent) == 0 {
		return true
	}

	switch object.(type) {
	case float64:
		f, err := strconv.ParseFloat(correspondent, 10)
		if err != nil {
			return false
		}

		return f == object
	}

	return fmt.Sprintf("%v", object) == correspondent
}

func NoErrors(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("failed to start test, %v", err)
	}
}

package mockutils

import "fmt"

func DoesObjectCorrespondToString(object any, correspondent string) bool {
	if object == nil && len(correspondent) == 0 {
		return true
	}

	return fmt.Sprintf("%v", object) == correspondent
}

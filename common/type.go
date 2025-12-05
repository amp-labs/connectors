// nolint:revive
package common

import "fmt"

//nolint:ireturn
func AssertType[T any](val any) (T, error) {
	of, ok := val.(T)
	if !ok {
		return of, fmt.Errorf("%w: expected type %T, but received %T", errFieldTypeMismatch, of, val)
	}

	return of, nil
}

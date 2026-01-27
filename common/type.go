// nolint:revive,godoclint
package common

import (
	"fmt"
)

//nolint:ireturn
func AssertType[T any](val any) (T, error) {
	of, ok := val.(T)
	if !ok {
		return of, fmt.Errorf("%w: expected type %T, but received %T", errFieldTypeMismatch, of, val)
	}

	return of, nil
}

func InferValueTypeFromData(value any) ValueType {
	if value == nil {
		return ValueTypeOther
	}

	switch value.(type) {
	case string:
		return ValueTypeString
	case float64, int, int64:
		return ValueTypeFloat
	case bool:
		return ValueTypeBoolean
	default:
		return ValueTypeOther
	}
}

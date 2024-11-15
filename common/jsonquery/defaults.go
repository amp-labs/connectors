package jsonquery

import (
	"errors"
	"strconv"
)

func (q *Query) IntegerWithDefault(key string, defaultValue int64) (int64, error) {
	result, err := q.Integer(key, true)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}

func (q *Query) StrWithDefault(key string, defaultValue string) (string, error) {
	result, err := q.Str(key, true)
	if err != nil {
		return "", err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}

// TextWithDefault returns a string under the key regardless of JSON data type.
// If data is not a string it will be converted to such.
func (q *Query) TextWithDefault(key string, defaultValue string) (string, error) {
	result, err := q.StrWithDefault(key, defaultValue)
	if err != nil {
		if !errors.Is(err, ErrNotString) {
			// Any error that is not due to non-string type is critical.
			return "", err
		}

		// Current data under `key` is not a string.
		// Explore other data types.
		// NOTE: as of now we check only if it is an integer.
		number, err := q.Integer(key, true)
		if err != nil {
			return "", err
		}

		if number == nil {
			return defaultValue, nil
		}

		return strconv.FormatInt(*number, 10), nil
	}

	return result, nil
}

func (q *Query) BoolWithDefault(key string, defaultValue bool) (bool, error) {
	result, err := q.Bool(key, true)
	if err != nil {
		return false, err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}

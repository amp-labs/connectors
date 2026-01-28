// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

var (
	errKeyNotFound       = errors.New("key not found")
	errNotANumber        = errors.New("not a number")
	errFieldTypeMismatch = errors.New("field type mismatch")
)

type StringMap map[string]any

func ToStringMap(m map[string]any) StringMap {
	return m
}

func (m StringMap) Keys() []string {
	return slices.AppendSeq(make([]string, 0, len(m)), maps.Keys(m))
}

func (m StringMap) Has(key string) bool {
	_, ok := m[key]

	return ok
}

func (m StringMap) Values() []any {
	return slices.AppendSeq(make([]any, 0, len(m)), maps.Values(m))
}

func (m StringMap) Len() int {
	return len(m)
}

func (m StringMap) GetCaseInsensitive(key string) (any, bool) {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}

	return nil, false
}

func (m StringMap) Get(key string) (any, error) {
	val, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errKeyNotFound, key)
	}

	return val, nil
}

// AsFloat is a helper function that extracts a float from the map.
// This function will convert number values such as int, uint, and float to float64.
// By doing so, it may lose precision for large numbers.
// This is helpful when you are not sure about the type of the number.
func (m StringMap) AsFloat(key string) (float64, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	t := reflect.TypeOf(val)

	//nolint:exhaustive
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(reflect.ValueOf(val).Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(reflect.ValueOf(val).Uint()), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(val).Float(), nil
	default:
		return 0, fmt.Errorf("%w: expected a number, but received %T", errNotANumber, val)
	}
}

// AsInt is a helper function that extracts an integer from the map.
// This function will convert number values such as int, uint, and float to int64.
// By doing so, it may lose precision for large numbers.
// This is helpful when you are not sure about the type of the number.
func (m StringMap) AsInt(key string) (int64, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	t := reflect.TypeOf(val)
	//nolint:exhaustive
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(val).Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(reflect.ValueOf(val).Uint()), nil //nolint:gosec
	case reflect.Float32, reflect.Float64:
		return int64(reflect.ValueOf(val).Float()), nil
	default:
		return 0, fmt.Errorf("%w: expected a number, but received %T", errNotANumber, val)
	}
}

func (m StringMap) GetString(key string) (string, error) {
	val, err := m.Get(key)
	if err != nil {
		return "", err
	}

	return AssertType[string](val)
}

func (m StringMap) GetBool(key string) (bool, error) {
	val, err := m.Get(key)
	if err != nil {
		return false, err
	}

	return AssertType[bool](val)
}

// GetInt extracts an integer from the map.
func (m StringMap) GetInt(key string) (int64, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	t := reflect.TypeOf(val)

	//nolint:exhaustive
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(val).Int(), nil
	default:
		return 0, fmt.Errorf("%w: expected an integer, but received %T", errFieldTypeMismatch, val)
	}
}

// GetFloat extracts a float from the map.
func (m StringMap) GetFloat(key string) (float64, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	t := reflect.TypeOf(val)

	//nolint:exhaustive
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(val).Float(), nil
	default:
		return 0, fmt.Errorf("%w: expected a float, but received %T", errFieldTypeMismatch, val)
	}
}

func (m StringMap) GetNumber(key string) (float64, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	t := reflect.TypeOf(val)

	//nolint:exhaustive
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(reflect.ValueOf(val).Int()), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(val).Float(), nil
	default:
		return 0, fmt.Errorf("%w: expected a number, but received %T", errFieldTypeMismatch, val)
	}
}

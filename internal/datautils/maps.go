// nolint:ireturn
package datautils

import (
	"encoding/json"

	"github.com/amp-labs/connectors/internal/goutils"
)

// Map is a generic version of map with useful methods.
// It can return Keys as a slice or a Set.
type Map[K comparable, V any] map[K]V

// FromMap converts golang map into Map resolving generic types on its own.
// Example:
//
//	Given:
//		dictionary = make(map[string]string)
//	Then statements are equivalent:
//		datautils.Map[string,string](golangMap)
//		datautils.FromMap(dictionary)
func FromMap[K comparable, V any](source map[K]V) Map[K, V] {
	return source
}

// ShallowCopy creates a shallow copy of the map.
// It copies the top-level keys and values, but does not clone
// nested or referenced objects. Use this when you only need
// a separate map container, not deep copies of the values.
func (m Map[K, V]) ShallowCopy() Map[K, V] {
	result := make(map[K]V)

	for key, value := range m {
		result[key] = value
	}

	return result
}

// DeepCopy creates a deep copy of the map.
// It duplicates all keys and values recursively using
// `goutils.Clone`, preserving type integrity for complex objects.
// Returns an error if cloning fails.
func (m Map[K, V]) DeepCopy() (Map[K, V], error) {
	return goutils.Clone(m)
}

func (m Map[K, V]) Keys() []K {
	keys := make([]K, 0)
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (m Map[K, V]) KeySet() Set[K] {
	return NewSetFromList(m.Keys())
}

func (m Map[K, V]) Has(key K) bool {
	_, ok := m[key]

	return ok
}

func (m Map[K, V]) Values() []V {
	values := make([]V, 0, len(m))

	for key := range m { // nolint:ireturn
		values = append(values, m[key])
	}

	return values
}

func (m Map[K, V]) AddMapValues(source Map[K, V]) {
	for k, v := range source {
		m[k] = v
	}
}

// Select returns the values for the given keys, along with the keys not found.
// It is the multi-key equivalent of a map lookup.
func (m Map[K, V]) Select(keys []K) ([]V, []K) {
	values := make([]V, 0)
	missingKeys := make([]K, 0)

	for _, key := range keys {
		value, ok := m[key]
		if !ok {
			missingKeys = append(missingKeys, key)
		} else {
			values = append(values, value)
		}
	}

	return values, missingKeys
}

// DefaultMap wrapper of the map that allows setting default return value on missing keys.
type DefaultMap[K comparable, V any] struct {
	// Map is a delegate.
	// All methods are embedded which grants the same capabilities, plus default value.
	Map[K, V]
	// When key is not found this callback will be used to provide default value.
	fallback func(key K) V
}

func NewDefaultMap[K comparable, V any](dict Map[K, V], fallback func(K) V) DefaultMap[K, V] {
	return DefaultMap[K, V]{
		Map:      dict,
		fallback: fallback,
	}
}

// Get method uses map with a fallback value.
// nolint:ireturn
func (m DefaultMap[K, V]) Get(key K) V {
	value, ok := m.Map[key]
	if ok {
		return value
	}

	if m.fallback != nil {
		return m.fallback(key)
	}

	var empty V

	return empty
}

// StructToMap convert a struct to a map of string to any.
func StructToMap(obj any) (map[string]any, error) {
	var result map[string]any

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

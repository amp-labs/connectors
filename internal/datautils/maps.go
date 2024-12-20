package datautils

import "encoding/json"

// Map is a generic version of map with useful methods.
// It can return Keys as a slice or a Set.
type Map[K comparable, V any] map[K]V

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

	for key := range m {
		values = append(values, m[key])
	}

	return values
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

// Convert a struct to a map of string to any.
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

package handy

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
	return NewSet(m.Keys())
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

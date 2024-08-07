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

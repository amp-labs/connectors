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

type Set[V comparable] map[V]struct{}

func NewSet[V comparable](values []V) Set[V] {
	result := make(Set[V])
	for _, v := range values {
		result[v] = struct{}{}
	}

	return result
}

func (s Set[V]) Diff(other Set[V]) []V {
	difference := s.Subtract(other)

	return append(difference, other.Subtract(s)...)
}

func (s Set[V]) Subtract(other Set[V]) []V {
	difference := make([]V, 0)

	for v := range s {
		if _, ok := other[v]; !ok {
			difference = append(difference, v)
		}
	}

	return difference
}

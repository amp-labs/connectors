package handy

// Set data structure that can hold any type of data.
type Set[T comparable] map[T]struct{}

// NewSet creates a set from slice.
func NewSet[V comparable](values []V) Set[V] {
	result := make(Set[V])
	result.Add(values)

	return result
}

// Add will add a list of values to the unique set. Repetitions will be omitted.
func (s Set[T]) Add(values []T) {
	for _, value := range values {
		s.AddOne(value)
	}
}

// AddOne will add single value to the unique set.
func (s Set[T]) AddOne(value T) {
	s[value] = struct{}{}
}

// List returns unique set in a shape of a slice.
func (s Set[T]) List() []T {
	list := make([]T, 0)
	for v := range s {
		list = append(list, v)
	}

	return list
}

// Diff is a difference between 2 sets.
// Every object that is not within intersection will be returned.
func (s Set[V]) Diff(other Set[V]) []V {
	difference := s.Subtract(other)

	return append(difference, other.Subtract(s)...)
}

// Subtract will return objects from current set that didn't occur in the input.
func (s Set[V]) Subtract(other Set[V]) []V {
	difference := make([]V, 0)

	for v := range s {
		if _, ok := other[v]; !ok {
			difference = append(difference, v)
		}
	}

	return difference
}

// Has returns true if key is found in the set.
func (s Set[T]) Has(key T) bool {
	_, ok := s[key]

	return ok
}

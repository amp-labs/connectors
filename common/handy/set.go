package handy

// NewStringSet and StringSet are Aliases.
var NewStringSet = NewSet[string] // nolint:gochecknoglobals
type StringSet = Set[string]

// Set data structure that can hold any type of data.
type Set[T comparable] map[T]struct{}

// NewSet creates a set from multiple values.
func NewSet[V comparable](values ...V) Set[V] {
	result := make(Set[V], len(values))
	result.Add(values)

	return result
}

// NewSetFromList creates a set from slice.
func NewSetFromList[V comparable](values []V) Set[V] {
	result := make(Set[V], len(values))
	result.Add(values)

	return result
}

func MergeSets[V comparable](first, second Set[V]) Set[V] {
	result := NewSet[V]()
	for element := range first {
		result.AddOne(element)
	}
	for element := range second {
		result.AddOne(element)
	}

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
	list := make([]T, len(s))
	index := 0

	for v := range s {
		list[index] = v
		index += 1
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

// Remove will delete a key from the set.
func (s Set[T]) Remove(key T) {
	delete(s, key)
}

func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

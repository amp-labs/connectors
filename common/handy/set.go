package handy

// Set data structure that can hold any type of data.
type Set[T comparable] map[T]bool

// Add will add a list of values to the unique set. Repetitions will be omitted.
func (s Set[T]) Add(values []T) {
	for _, value := range values {
		s.AddOne(value)
	}
}

// AddOne will add single value to the unique set.
func (s Set[T]) AddOne(value T) {
	s[value] = true
}

// List returns unique set in a shape of slice.
func (s Set[T]) List() []T {
	list := make([]T, 0)
	for v := range s {
		list = append(list, v)
	}

	return list
}

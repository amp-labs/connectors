package handy

type Set[T comparable] map[T]bool

func (s Set[T]) Add(values []T) {
	for _, value := range values {
		s.AddOne(value)
	}
}

func (s Set[T]) AddOne(value T) {
	s[value] = true
}

func (s Set[T]) List() []T {
	list := make([]T, 0)
	for v := range s {
		list = append(list, v)
	}

	return list
}

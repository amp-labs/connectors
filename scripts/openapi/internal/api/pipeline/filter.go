package pipeline

type FilterFunc[T any] func(T) bool

func (p Pipeline[T]) Filter(fn FilterFunc[T]) Pipeline[T] {
	items := make([]T, 0, len(p.items))

	for _, item := range p.items {
		if fn(item) {
			items = append(items, item)
		}
	}

	return Pipeline[T]{items: items}
}

func And[T any](fns ...FilterFunc[T]) FilterFunc[T] {
	return func(t T) bool {
		for _, fn := range fns {
			if !fn(t) {
				return false
			}
		}

		return true
	}
}

func Or[T any](fns ...FilterFunc[T]) FilterFunc[T] {
	return func(t T) bool {
		for _, fn := range fns {
			if fn(t) {
				return true
			}
		}

		return false
	}
}

func Not[T any](fn FilterFunc[T]) FilterFunc[T] {
	return func(t T) bool { return !fn(t) }
}

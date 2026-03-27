package pipeline

type MapFunc[T any] func(T) T

func (p Pipeline[T]) Map(fn MapFunc[T]) Pipeline[T] {
	out := make([]T, len(p.items))
	for i, item := range p.items {
		out[i] = fn(item)
	}

	return Pipeline[T]{items: out}
}

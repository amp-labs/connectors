package pipeline

type ReduceFunc[T any] func([]T) []T

func (p Pipeline[T]) Reduce(resolver ReduceFunc[T]) Pipeline[T] {
	return Pipeline[T]{items: resolver(p.items)}
}

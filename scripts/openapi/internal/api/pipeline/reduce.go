package pipeline

type ReduceFunc[FROM, TO any] func([]FROM) []TO

func (p Pipeline[FROM]) Reduce[TO any](resolver ReduceFunc[FROM, TO]) Pipeline[TO] {
	return Pipeline[TO]{items: resolver(p.items)}
}

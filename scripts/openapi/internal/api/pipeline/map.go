package pipeline

type MapFunc[FROM, TO any] func(FROM) TO

func (p Pipeline[FROM]) Map[TO any](fn MapFunc[FROM, TO]) Pipeline[TO] {
	out := make([]TO, len(p.items))
	for i, item := range p.items {
		out[i] = fn(item)
	}

	return Pipeline[TO]{items: out}
}

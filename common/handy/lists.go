package handy

type Lists[T any] map[string][]T

func (l Lists[T]) Add(kind string, objects ...T) {
	if _, ok := l[kind]; !ok {
		l[kind] = make([]T, 0)
	}

	l[kind] = append(l[kind], objects...)
}

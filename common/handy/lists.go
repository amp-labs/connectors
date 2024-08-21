package handy

// Lists is a collection of named lists.
type Lists[T any] map[string][]T

// Add objects into the named list.
// If the list didn't exist it will create it.
func (l Lists[T]) Add(kind string, objects ...T) {
	if _, ok := l[kind]; !ok {
		l[kind] = make([]T, 0)
	}

	l[kind] = append(l[kind], objects...)
}

// GetBucketNames returns names of lists, buckets they are in.
// Example:
//
//	veggies: cucumber, tomato;
//	fruits: pineapple;
//
// Will return veggies and fruits.
func (l Lists[T]) GetBucketNames() []string {
	result := make([]string, len(l))

	index := 0

	for name := range l {
		result[index] = name
		index++
	}

	return result
}

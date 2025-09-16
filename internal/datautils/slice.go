package datautils

func EqualUnordered[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	count := make(map[T]int)
	for _, v := range a {
		count[v]++
	}

	for _, v := range b {
		count[v]--
		if count[v] < 0 {
			return false
		}
	}

	return true
}

// SliceToMap builds a map from a slice, using makeKey to derive each entry's key.
func SliceToMap[K comparable, V any](list []V, makeKey func(V) K) Map[K, V] {
	result := make(map[K]V)

	for _, value := range list {
		key := makeKey(value)
		result[key] = value
	}

	return result
}

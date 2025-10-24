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

// ForEach applies the provided function f to each element of s,
// returning a new slice containing the results.
//
// Example:
//
//	names := []string{"alice", "bob"}
//	lengths := datautils.Map(names, func(s string) int { return len(s) })
//	// lengths == []int{5, 3}
func ForEach[F, T any](input []F, mapper func(F) T) []T {
	if len(input) == 0 {
		return make([]T, 0)
	}

	output := make([]T, len(input))
	for index, value := range input {
		output[index] = mapper(value)
	}

	return output
}

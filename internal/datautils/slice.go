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

// ToAnySlice converts a slice of any type T to a slice of empty interface values ([]any).
// This is useful when you need to pass a typed slice to functions that expect []any.
//
// Example:
//
//	ints := []int{1, 2, 3}
//	anySlice := datautils.ToAnySlice(ints)
//	// anySlice == []any{1, 2, 3}
func ToAnySlice[T any](slice []T) []any {
	result := make([]any, len(slice))

	for i, v := range slice {
		result[i] = v
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

// ForEachWithErr applies the provided mapper function to each element of the input slice.
// If the mapper returns an error for any element, the function returns immediately with that error.
// Otherwise, it returns a new slice containing the mapped results.
func ForEachWithErr[F, T any](input []F, mapper func(F) (T, error)) ([]T, error) {
	if len(input) == 0 {
		return make([]T, 0), nil
	}

	var (
		err    error
		output = make([]T, len(input))
	)

	for index, value := range input {
		output[index], err = mapper(value)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

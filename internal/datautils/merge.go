package datautils

// MergeIndexedLists combines IndexedLists into one.
func MergeIndexedLists[ID comparable, V any](lists ...IndexedLists[ID, V]) IndexedLists[ID, V] {
	result := make(IndexedLists[ID, V])

	for _, collection := range lists {
		for key, value := range collection {
			result.Add(key, value...)
		}
	}

	return result
}

// MergeNamedLists combines NamedLists into one.
func MergeNamedLists[V any](lists ...NamedLists[V]) NamedLists[V] {
	result := make(NamedLists[V])

	for _, collection := range lists {
		for key, value := range collection {
			result.Add(key, value...)
		}
	}

	return result
}

// MergeUniqueLists combines UniqueLists into one.
func MergeUniqueLists[ID, V comparable](lists ...UniqueLists[ID, V]) UniqueLists[ID, V] {
	result := make(UniqueLists[ID, V])

	for _, collection := range lists {
		for key, value := range collection {
			result.Add(key, value.List()...)
		}
	}

	return result
}

func MergeSets[T comparable](sets ...Set[T]) Set[T] {
	result := NewSet[T]()

	for _, set := range sets {
		for value := range set {
			result.AddOne(value)
		}
	}

	return result
}

// MergeSlices combines multiple slices of the same type into a single slice.
func MergeSlices[T any](slices ...[]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}

package pipeline

import "github.com/amp-labs/connectors/internal/datautils"

func Combine[T any](
	left, right Pipeline[T],
	keyFn func(T) string,
	mergeFn func(left, right T) T,
) Pipeline[T] {
	registry := datautils.Map[string, T]{}

	// Add items from left.
	for _, value := range left.List() {
		key := keyFn(value)
		registry[key] = value
	}

	// Add items from right.
	// If already present ask mergeFn to break the tie.
	for _, value := range right.List() {
		key := keyFn(value)
		if existing, ok := registry[key]; ok {
			registry[key] = mergeFn(existing, value)
		} else {
			registry[key] = value
		}
	}

	return New[T](registry.Values())
}

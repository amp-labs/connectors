package goutils

// Pointer returns a pointer to the given value.
func Pointer[T any](input T) *T {
	return &input
}

// MustBeNil - panics on non-empty error.
func MustBeNil(err error) {
	if err != nil {
		panic(err)
	}
}

// PanicRecovery catches the cause of panic and passes it to the callback.
func PanicRecovery(wrapup func(cause error)) {
	if re := recover(); re != nil {
		err, ok := re.(error)
		if !ok {
			panic(re)
		}

		wrapup(err)
	}
}

// Identity returns the input value unchanged.
// It acts as an identity function for any type.
func Identity[T any](input T) T {
	return input
}

// ToAnySlice converts a slice of any type T into a slice of empty interfaces ([]any).
// It preserves the length and order of the input slice, boxing each element into an
// interface value. If the input slice is nil, the result is also nil.
//
// Example:
//
//	nums := []int{1, 2, 3}
//	anySlice := ToAnySlice(nums) // []any{1, 2, 3}
func ToAnySlice[T any](list []T) []any {
	if list == nil {
		return nil
	}

	result := make([]any, len(list))
	for index, element := range list {
		result[index] = element
	}

	return result
}

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

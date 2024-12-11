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

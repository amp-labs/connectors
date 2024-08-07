package handy

// Pointers is a collection of utility functions that help with go pointer referencing.
var Pointers = pointers{} // nolint:gochecknoglobals

type pointers struct{}

// Str returns a pointer to the given string.
func (pointers) Str(input string) *string {
	return &input
}

// Bool returns a pointer to the given boolean.
func (pointers) Bool(input bool) *bool {
	return &input
}

// IsTrue checks boolean pointer was set and is true.
func (pointers) IsTrue(input *bool) bool {
	if input == nil {
		return false
	}

	return *input
}

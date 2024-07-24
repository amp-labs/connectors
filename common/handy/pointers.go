package handy

var Pointers = pointers{} // nolint:gochecknoglobals

type pointers struct{}

func (pointers) Str(input string) *string {
	return &input
}

func (pointers) Bool(input bool) *bool {
	return &input
}

func (pointers) IsTrue(input *bool) bool {
	if input == nil {
		return false
	}

	return *input
}

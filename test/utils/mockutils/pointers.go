package mockutils

var Pointers = pointers{}

type pointers struct{}

func (pointers) Str(input string) *string {
	return &input
}

func (pointers) Bool(input bool) *bool {
	return &input
}

package interpreter

// ErrorDescriptor enhances base error with extra message.
// Every implementor decides how server response will be converted, and
// how important message will be formatted into helpful error.
type ErrorDescriptor interface {
	CombineErr(base error) error
}

// FormatTemplate holds concrete struct that represent erroneous server response.
// It is used by FormatSwitch.
type FormatTemplate struct {
	// MustKeys is a list of important keys that if all present will signify the match for Template.
	MustKeys []string
	// Template is a factory that returns a struct, which will be used to flush the data into.
	Template Template
}

type Template func() ErrorDescriptor

type Templates []Template

// when all required keys are present in the payload it returns true.
func (t FormatTemplate) matches(payload map[string]any) bool {
	for _, pivot := range t.MustKeys {
		if _, ok := payload[pivot]; !ok {
			return false
		}
	}

	// Every key is present.
	// Empty list means instant match.
	return true
}

package interpreter

// FormatTemplate holds concrete struct that represent erroneous server response.
// It is used by FormatSwitch.
type FormatTemplate struct {
	// MustKeys is a list of important keys that if all present will signify the match for Template.
	MustKeys []string
	// Template is a struct pointer which will be used to flush the data into.
	// Must implement common.ErrorDescriptor.
	Template any
}

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

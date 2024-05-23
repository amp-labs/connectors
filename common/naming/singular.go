package naming

import "github.com/gertd/go-pluralize"

var pluralizer = pluralize.NewClient() // nolint:gochecknoglobals

type SingularString struct {
	text string
}
type SingularStrings []SingularString

func NewSingularString(str string) SingularString {
	return SingularString{text: pluralizer.Singular(str)}
}

func (s SingularString) String() string {
	return s.text
}

func (s SingularString) Plural() PluralString {
	return NewPluralString(s.String())
}

func NewSingularStrings(list []string) SingularStrings {
	result := make(SingularStrings, len(list))
	for i, str := range list {
		result[i] = NewSingularString(str)
	}

	return result
}

func (s SingularStrings) Plural() PluralStrings {
	result := make(PluralStrings, len(s))
	for i, str := range s {
		result[i] = str.Plural()
	}

	return result
}

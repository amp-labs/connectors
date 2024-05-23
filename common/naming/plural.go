package naming

type PluralString struct {
	text string
}
type PluralStrings []PluralString

func NewPluralString(str string) PluralString {
	return PluralString{text: pluralizer.Plural(str)}
}

func (s PluralString) String() string {
	return s.text
}

func (s PluralString) Singular() SingularString {
	return NewSingularString(s.String())
}

func NewPluralStrings(list []string) PluralStrings {
	result := make(PluralStrings, len(list))
	for i, str := range list {
		result[i] = NewPluralString(str)
	}

	return result
}

func (s PluralStrings) Plural() SingularStrings {
	result := make(SingularStrings, len(s))
	for i, str := range s {
		result[i] = str.Singular()
	}

	return result
}

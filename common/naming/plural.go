package naming

import "encoding/json"

// PluralString imposes plural form on a word conforming to English rules.
// It is capable of self conversion to singular form.
// You can use it as keys in maps, values, and it knows how to Marshal itself like a string.
// Unmarshalling will apply plural formating.
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

func (s PluralStrings) Singular() SingularStrings {
	result := make(SingularStrings, len(s))
	for i, str := range s {
		result[i] = str.Singular()
	}

	return result
}

func (s PluralString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.text)
}

func (s *PluralString) UnmarshalJSON(bytes []byte) error {
	var text string

	err := json.Unmarshal(bytes, &text)
	if err != nil {
		return err
	}

	s.text = pluralizer.Plural(text)

	return nil
}

func (s PluralString) MarshalText() ([]byte, error) {
	return []byte(s.text), nil
}

func (s *PluralString) UnmarshalText(text []byte) error {
	return s.UnmarshalJSON(text)
}

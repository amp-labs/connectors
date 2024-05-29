package naming

import (
	"encoding/json"

	"github.com/gertd/go-pluralize"
)

// This client guides the behaviour of this package and sets all rules for SingularString and PluralString.
var pluralizer = pluralize.NewClient() // nolint:gochecknoglobals

// SingularString imposes singular form on a word conforming to English rules.
// It is capable of self conversion to plural form.
// You can use it as keys in maps, values, and it knows how to Marshal itself like a string.
// Unmarshalling will apply singular formating.
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

func (s SingularString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.text)
}

func (s *SingularString) UnmarshalJSON(bytes []byte) error {
	var text string

	err := json.Unmarshal(bytes, &text)
	if err != nil {
		return err
	}

	s.text = pluralizer.Singular(text)

	return nil
}

func (s SingularString) MarshalText() ([]byte, error) {
	return []byte(s.text), nil
}

func (s *SingularString) UnmarshalText(text []byte) error {
	return s.UnmarshalJSON(text)
}

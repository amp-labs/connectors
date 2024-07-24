package naming

import (
	"encoding/json"
	"strings"
)

type LowerString struct {
	text string
}

func NewLowerString(str string) LowerString {
	return LowerString{text: strings.ToLower(str)}
}

func (s LowerString) String() string {
	return s.text
}

func (s LowerString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.text)
}

func (s *LowerString) UnmarshalJSON(bytes []byte) error {
	var text string

	err := json.Unmarshal(bytes, &text)
	if err != nil {
		return err
	}

	s.text = NewLowerString(text).String()

	return nil
}

func (s LowerString) MarshalText() ([]byte, error) {
	return []byte(s.text), nil
}

func (s *LowerString) UnmarshalText(text []byte) error {
	return s.UnmarshalJSON(text)
}

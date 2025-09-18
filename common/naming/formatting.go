package naming

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CapitalizeFirstLetterEveryWord(text string) string {
	caser := cases.Title(language.English)

	text = caser.String(text)
	for from, to := range map[string]string{
		" For ": " for ",
		" A ":   " a ",
	} {
		text = strings.ReplaceAll(text, from, to)
	}

	return text
}

func CapitalizeFirstLetter(text string) string {
	if len(text) == 0 {
		return text
	}

	caser := cases.Title(language.English)

	return caser.String(text[:1]) + text[1:]
}

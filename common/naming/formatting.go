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

func PluralityAndCaseIgnoreEqual(a, b string) bool {
	singularA := pluralizer.Singular(a)
	singularB := pluralizer.Singular(b)
	pluralA := pluralizer.Plural(a)
	pluralB := pluralizer.Plural(b)

	return strings.EqualFold(singularA, singularB) && strings.EqualFold(pluralA, pluralB)
}

package naming

import (
	"strings"
	"unicode"

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

// SeparateUnderscoreWords converts underscore-separated words into space-separated words.
func SeparateUnderscoreWords(text string) string {
	return strings.ReplaceAll(text, "_", " ")
}

func SeparateDotWords(text string) string {
	return strings.ReplaceAll(text, ".", " ")
}

// SeparateCamelCaseWords converts camelCase or PascalCase into space-separated words
// by inserting a space before each uppercase letter that follows a lowercase letter.
// Acronyms (consecutive uppercase letters) are kept as a single word.
func SeparateCamelCaseWords(text string) string {
	var b strings.Builder

	var prev rune
	for i, r := range text {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(prev) {
			b.WriteRune(' ')
		}

		b.WriteRune(r)
		prev = r
	}

	return b.String()
}

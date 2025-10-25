package naming

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts a string to snake_case.
// Handles PascalCase, camelCase, and existing snake_case inputs.
// Examples:
//   - "UserId" -> "user_id"
//   - "EventType" -> "event_type"
//   - "HTTPResponse" -> "http_response"
//   - "user_id" -> "user_id" (already snake_case)
func ToSnakeCase(input string) string {
	if input == "" {
		return input
	}

	// If already in snake_case (lowercase with underscores), return as-is
	if isSnakeCase(input) {
		return input
	}

	const extraCapacity = 5

	var result strings.Builder

	result.Grow(len(input) + extraCapacity) // Allocate extra space for potential underscores

	runes := []rune(input)
	for i, r := range runes {
		// Handle uppercase letters
		if unicode.IsUpper(r) {
			// Add underscore before uppercase if not at start
			if i > 0 && shouldAddUnderscore(runes, i) {
				result.WriteRune('_')
			}

			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToCamelCase converts a string to camelCase (first letter lowercase).
// Handles various input formats: PascalCase, snake_case, kebab-case, space-separated.
// Examples:
//   - "first_name" -> "firstName"
//   - "FirstName" -> "firstName"
//   - "is-archived" -> "isArchived"
func ToCamelCase(input string) string {
	if input == "" {
		return input
	}

	// Handle already camelCase strings (common case)
	if !strings.ContainsAny(input, "_ -") && unicode.IsLower(rune(input[0])) {
		return input
	}

	words := splitIntoWords(input)
	if len(words) == 0 {
		return input
	}

	// First word is lowercase, rest are title-cased (first letter upper, rest lower)
	result := strings.ToLower(words[0])

	for i := 1; i < len(words); i++ {
		if words[i] != "" {
			runes := []rune(words[i])
			result += string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
		}
	}

	return result
}

// ToPascalCase converts a string to PascalCase (first letter uppercase).
// Handles various input formats: camelCase, snake_case, kebab-case, space-separated.
// Examples:
//   - "user_id" -> "UserId"
//   - "firstName" -> "FirstName"
//   - "email-address" -> "EmailAddress"
func ToPascalCase(input string) string {
	if input == "" {
		return input
	}

	// Handle snake_case or kebab-case
	if strings.ContainsAny(input, "_ -") {
		words := splitIntoWords(input)

		var result strings.Builder

		for _, word := range words {
			if word != "" {
				runes := []rune(word)
				result.WriteRune(unicode.ToUpper(runes[0]))

				if len(runes) > 1 {
					result.WriteString(strings.ToLower(string(runes[1:])))
				}
			}
		}

		return result.String()
	}

	// For camelCase or other formats, just capitalize the first letter but preserve rest
	return CapitalizeFirstOnly(input)
}

// capitalizeFirst capitalizes the first letter of a string and lowercases the rest.
func capitalizeFirst(input string) string {
	if input == "" {
		return input
	}

	runes := []rune(input)

	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}

// CapitalizeFirstOnly capitalizes only the first letter without changing the rest.
func CapitalizeFirstOnly(input string) string {
	if input == "" {
		return input
	}

	runes := []rune(input)

	return string(unicode.ToUpper(runes[0])) + string(runes[1:])
}

// ToLowerCase converts a string to lowercase.
// This is a convenience wrapper around strings.ToLower for consistency with other case functions.
func ToLowerCase(input string) string {
	return strings.ToLower(input)
}

// splitIntoWords splits a string into words based on delimiters or case changes.
func splitIntoWords(input string) []string {
	if input == "" {
		return nil
	}

	// Handle delimiter-based splitting (snake_case, kebab-case, space-separated)
	if strings.ContainsAny(input, "_ -") {
		normalized := strings.ReplaceAll(input, "-", " ")
		normalized = strings.ReplaceAll(normalized, "_", " ")
		parts := strings.Fields(normalized)

		words := make([]string, 0, len(parts))

		for _, part := range parts {
			if part != "" {
				words = append(words, part)
			}
		}

		return words
	}

	// Handle PascalCase/camelCase by splitting on capital letters
	return splitOnCaps(input)
}

// splitOnCaps splits a string on capital letters (handles PascalCase/camelCase).
func splitOnCaps(input string) []string {
	if input == "" {
		return nil
	}

	var words []string

	var currentWord strings.Builder

	runes := []rune(input)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			// Start new word on capital letter
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}

		currentWord.WriteRune(r)
	}

	// Add final word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// isSnakeCase checks if a string is already in snake_case format.
func isSnakeCase(input string) bool {
	// Must be lowercase and can contain underscores
	for _, r := range input {
		if unicode.IsUpper(r) {
			return false
		}
	}

	return true
}

// shouldAddUnderscore determines if an underscore should be added before the current character.
// Handles various edge cases like consecutive uppercase letters (HTTPResponse -> http_response).
func shouldAddUnderscore(runes []rune, idx int) bool {
	if idx <= 0 || idx >= len(runes) {
		return false
	}

	current := runes[idx]
	prev := runes[idx-1]

	// Don't add if previous char is already underscore
	if prev == '_' {
		return false
	}

	// Add if previous char is lowercase or digit (transition from lower to upper)
	if unicode.IsLower(prev) || unicode.IsDigit(prev) {
		return true
	}

	// For uppercase sequences, add underscore before the last uppercase
	// if it's followed by lowercase (e.g., "HTTPResponse" -> "http_response")
	if unicode.IsUpper(prev) && unicode.IsUpper(current) {
		if idx+1 < len(runes) && unicode.IsLower(runes[idx+1]) {
			return true
		}
	}

	return false
}

// IsUpper checks if a rune is an uppercase letter.
func IsUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// IsLower checks if a rune is a lowercase letter.
func IsLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// ToLowerRune converts a rune to lowercase using ASCII offset.
func ToLowerRune(r rune) rune {
	if IsUpper(r) {
		return r + ('a' - 'A')
	}

	return r
}

// ToUpperRune converts a rune to uppercase using ASCII offset.
func ToUpperRune(r rune) rune {
	if IsLower(r) {
		return r - ('a' - 'A')
	}

	return r
}

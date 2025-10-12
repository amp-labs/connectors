package ashby

import (
	"context"
	"strings"
	"unicode"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Ashby naming conventions.
// Ashby uses camelCase for both objects and fields, with objects in singular form.
//
// Objects:
//   - Converts to singular camelCase (first letter lowercase)
//   - Examples: "Application" -> "application", "Candidates" -> "candidate"
//   - Compound names: "InterviewSchedule" -> "interviewSchedule", "job_posting" -> "jobPosting"
//
// Fields:
//   - Converts to camelCase (first letter lowercase)
//   - Examples: "FirstName" -> "firstName", "is_archived" -> "isArchived"
//
// Note: Ashby API is case-sensitive and requires exact camelCase matching.
// The API uses dot notation for operations (e.g., "candidate.list", "application.create").
func (c *Connector) NormalizeEntityName(
	ctx context.Context, entity connectors.Entity, input string,
) (normalized string, err error) {
	switch entity {
	case connectors.EntityObject:
		return normalizeObjectName(input), nil
	case connectors.EntityField:
		return normalizeFieldName(input), nil
	default:
		// Unknown entity type, return unchanged
		return input, nil
	}
}

// normalizeObjectName converts object names to singular camelCase.
// Ashby's standard objects are singular and camelCase: application, candidate, job, user.
func normalizeObjectName(input string) string {
	// Convert to singular form
	singular := naming.NewSingularString(input).String()

	// Convert to camelCase
	return toCamelCase(singular)
}

// normalizeFieldName converts field names to camelCase.
// Ashby field names use camelCase: firstName, isArchived, primaryEmailAddress.
func normalizeFieldName(input string) string {
	return toCamelCase(input)
}

// toCamelCase converts a string to camelCase (first letter lowercase).
// Handles various input formats: PascalCase, snake_case, kebab-case, space-separated.
func toCamelCase(input string) string {
	if input == "" {
		return input
	}

	// Handle already camelCase strings (common case)
	if !strings.ContainsAny(input, "_ -") && !unicode.IsUpper(rune(input[0])) {
		return input
	}

	// Split on common delimiters: underscore, dash, space
	var words []string

	if strings.ContainsAny(input, "_ -") {
		// Split on delimiters
		normalized := strings.ReplaceAll(input, "-", " ")
		normalized = strings.ReplaceAll(normalized, "_", " ")
		parts := strings.Fields(normalized)
		words = make([]string, 0, len(parts))

		for _, part := range parts {
			if part != "" {
				words = append(words, part)
			}
		}
	} else {
		// Handle PascalCase or camelCase - split on capital letters
		words = splitOnCaps(input)
	}

	if len(words) == 0 {
		return input
	}

	// First word is lowercase, rest are capitalized
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += capitalizeFirst(words[i])
	}

	return result
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

// capitalizeFirst capitalizes the first letter of a string and lowercases the rest.
func capitalizeFirst(input string) string {
	if input == "" {
		return input
	}

	runes := []rune(input)

	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}

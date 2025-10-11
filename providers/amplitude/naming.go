package amplitude

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Amplitude naming conventions.
// Amplitude uses lowercase with underscores (snake_case) for both objects and fields.
// The API is case-sensitive and requires exact matching.
//
// Objects:
//   - Converts to lowercase plural form with underscores
//   - Preserves special characters like slashes (/) and hyphens (-) in taxonomy paths
//   - Examples: "Events" -> "events", "Cohort" -> "cohorts", "TaxonomyEvent" -> "taxonomy/event"
//   - Taxonomy objects: "taxonomy/category", "taxonomy/event", "taxonomy/event-property",
//     "taxonomy/user-property", "taxonomy/group-property"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "UserId" -> "user_id", "EventType" -> "event_type", "SessionID" -> "session_id"
//
// Special Cases:
//   - Taxonomy paths with "/" are preserved (e.g., "taxonomy/event")
//   - Hyphens in property names are preserved (e.g., "event-property")
//   - lookup_table uses underscore not hyphen
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

// normalizeObjectName converts object names to lowercase plural form.
// Amplitude's standard objects use lowercase plural or lowercase singular for conceptual nouns.
// Special handling for taxonomy paths which include "/" separators and "-" in property names.
func normalizeObjectName(input string) string {
	// Handle empty input
	if input == "" {
		return input
	}

	// Special case: taxonomy paths with slashes (taxonomy/event, taxonomy/event-property, etc.)
	// These should remain as-is but lowercased
	if strings.Contains(input, "/") {
		return strings.ToLower(input)
	}

	// Special case: if the input already has underscores or hyphens, preserve them
	// Examples: lookup_table, event-property
	if strings.Contains(input, "_") || strings.Contains(input, "-") {
		return strings.ToLower(input)
	}

	// For simple names, convert to plural and lowercase
	// Examples: event -> events, cohort -> cohorts, annotation -> annotations
	plural := naming.NewPluralString(input).String()

	return strings.ToLower(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Amplitude field names use snake_case convention (user_id, event_type, etc.)
func normalizeFieldName(input string) string {
	// Handle empty input
	if input == "" {
		return input
	}

	// If already in snake_case (contains underscores), just lowercase it
	if strings.Contains(input, "_") {
		return strings.ToLower(input)
	}

	// Convert camelCase or PascalCase to snake_case
	// This handles cases like "UserId" -> "user_id", "EventType" -> "event_type"
	return toSnakeCase(input)
}

// toSnakeCase converts a camelCase or PascalCase string to snake_case.
// Examples: "UserId" -> "user_id", "EventType" -> "event_type", "SessionID" -> "session_id".
func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	const extraCapacity = 5

	var result strings.Builder

	result.Grow(len(input) + extraCapacity) // Allocate extra space for potential underscores

	for idx, char := range input {
		// If this is an uppercase letter and not the first character
		if idx > 0 && char >= 'A' && char <= 'Z' {
			// Add underscore before uppercase letter
			// Special handling: don't add underscore if previous char was also uppercase
			// (handles acronyms like "ID" -> "id" instead of "i_d")
			prevChar := rune(input[idx-1])
			if prevChar < 'A' || prevChar > 'Z' {
				result.WriteRune('_')
			}
		}

		result.WriteRune(char)
	}

	return strings.ToLower(result.String())
}

package avoma

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Avoma naming conventions.
// Avoma uses lowercase snake_case for both objects and fields. Most objects are plural,
// with the notable exception of "template" which is singular.
//
// Objects:
//   - Converts to lowercase plural with underscores for multi-word names
//   - Special case: "template" remains singular
//   - Examples: "Call" -> "calls", "Meeting" -> "meetings", "CustomCategory" -> "custom_categories"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "ExternalId" -> "external_id", "StartAt" -> "start_at"
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

// normalizeObjectName converts object names to lowercase plural with underscores.
// Avoma's standard objects follow this pattern: calls, meetings, custom_categories, etc.
// Special case: "template" is singular in the API.
func normalizeObjectName(input string) string {
	// Convert to snake_case first (handles camelCase and PascalCase)
	snakeCase := toSnakeCase(input)

	// Check for the special singular case
	if strings.EqualFold(snakeCase, "template") {
		return "template"
	}

	// For all other objects, convert to plural
	plural := naming.NewPluralString(snakeCase).String()

	return strings.ToLower(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Avoma field names use snake_case: external_id, start_at, is_voicemail, etc.
func normalizeFieldName(input string) string {
	return toSnakeCase(input)
}

const (
	// bufferSize is the extra capacity for underscores when converting to snake_case.
	bufferSize = 5
)

// toSnakeCase converts a string to snake_case.
// Handles both camelCase and PascalCase inputs.
func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder

	result.Grow(len(input) + bufferSize) // Pre-allocate with some buffer for underscores

	for i, r := range input {
		// If uppercase and not at the start, add underscore before it
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Don't add underscore if previous char was already underscore
			if input[i-1] != '_' {
				result.WriteRune('_')
			}
		}

		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

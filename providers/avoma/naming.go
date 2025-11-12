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
	snakeCase := naming.ToSnakeCase(input)

	// Check for the special singular case
	if strings.EqualFold(snakeCase, "template") {
		return "template"
	}

	// For all other objects, convert to plural and lowercase
	plural := naming.NewPluralString(snakeCase).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Avoma field names use snake_case: external_id, start_at, is_voicemail, etc.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

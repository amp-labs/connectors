package dixa

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Dixa naming conventions.
// Dixa uses lowercase plural for objects and camelCase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Preserves hyphens and slashes in compound names
//   - Examples: "Agent" -> "agents", "EndUser" -> "endusers", "Queue" -> "queues"
//   - Compound examples: "custom-attributes", "contact-endpoints", "business-hours/schedules"
//
// Fields:
//   - Converts to camelCase (lowercase first letter)
//   - Examples: "FirstName" -> "firstName", "DisplayName" -> "displayName", "ID" -> "id"
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

// normalizeObjectName converts object names to lowercase plural.
// Dixa's standard objects are always lowercase plural: agents, endusers, queues, teams, tags.
// Some objects contain hyphens (custom-attributes) or slashes (business-hours/schedules).
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to camelCase.
// Dixa field names use camelCase with lowercase first letter (firstName, lastName, displayName).
func normalizeFieldName(input string) string {
	return naming.ToCamelCase(input)
}

package copper

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Copper CRM naming conventions.
// Copper uses lowercase with underscores (snake_case) for both objects and fields.
//
// Objects:
//   - Converts to lowercase plural form with underscores
//   - Examples: "Company" -> "companies", "ActivityType" -> "activity_types", "People" -> "people"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "FirstName" -> "first_name", "Email" -> "email", "DateCreated" -> "date_created"
//
// Custom Fields:
//   - Custom fields are prefixed with "custom_field_" in the connector layer
//   - Example: "Birthday" -> "custom_field_birthday"
//
// Reference: https://developer.copper.com/
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
// Copper objects are always lowercase plural and use underscores for multi-word names.
// Examples: "Company" -> "companies", "ActivityType" -> "activity_types", "Person" -> "people".
func normalizeObjectName(input string) string {
	// Convert to plural form first
	plural := naming.NewPluralString(input).String()

	// Convert to snake_case (lowercase with underscores)
	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase with underscores (snake_case).
// Copper field names consistently use snake_case.
// Examples: "FirstName" -> "first_name", "DateCreated" -> "date_created", "Email" -> "email".
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

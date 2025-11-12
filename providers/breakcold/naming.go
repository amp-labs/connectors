package breakcold

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Breakcold naming conventions.
// Breakcold uses lowercase plural for objects and snake_case for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Lead" -> "leads", "Status" -> "status", "Reminder" -> "reminders"
//
// Fields:
//   - Converts to lowercase (already snake_case in API, but normalized for consistency)
//   - Examples: "FirstName" -> "firstname", "first_name" -> "first_name", "Email" -> "email"
//
// Note: Breakcold API has some endpoint inconsistencies:
//   - Some read endpoints use /objectname/list (e.g., /leads/list, /reminders/list)
//   - POST operations may use singular names (e.g., /lead, /attribute)
//   - PATCH/DELETE operations use plural names (e.g., /leads/{id}, /attributes/{id})
//
// This normalizer standardizes to lowercase plural for consistency, and the connector
// handles the specific endpoint variations in its request builders.
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
// Breakcold's API uses lowercase plural object names in most endpoints.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to lowercase.
// Breakcold uses snake_case for all fields (e.g., first_name, linkedin_url).
// This normalizer converts any input to lowercase to ensure consistency.
func normalizeFieldName(input string) string {
	return naming.ToLowerCase(input)
}

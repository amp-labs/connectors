package closecrm

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Close CRM naming conventions.
// Close CRM uses lowercase singular for objects and snake_case lowercase for fields.
//
// Objects:
//   - Converts to lowercase singular form
//   - Examples: "Lead" -> "lead", "Contacts" -> "contact", "Activity" -> "activity"
//
// Fields:
//   - Converts to lowercase (Close uses snake_case)
//   - Preserves dots for custom fields (custom.cf_*)
//   - Examples: "DateCreated" -> "datecreated", "user_name" -> "user_name"
//   - Note: Snake_case conversion is handled by the API, we just lowercase
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

// normalizeObjectName converts object names to lowercase singular.
// Close CRM objects are singular: lead, activity, contact, opportunity, task.
func normalizeObjectName(input string) string {
	// Convert to singular form and lowercase
	singular := naming.NewSingularString(input).String()

	return naming.ToLowerCase(singular)
}

// normalizeFieldName converts field names to lowercase.
// Close CRM field names are snake_case and lowercase.
// Custom fields use dot notation: custom.cf_<ID>.
func normalizeFieldName(input string) string {
	return strings.ToLower(input)
}

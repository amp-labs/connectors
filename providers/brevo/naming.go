package brevo

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Brevo naming conventions.
// Brevo uses lowercase plural for objects and camelCase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Contact" -> "contacts", "Company" -> "companies", "Deal" -> "deals"
//
// Fields:
//   - Converts to lowercase
//   - Examples: "FirstName" -> "firstname", "Email" -> "email", "CreatedAt" -> "createdat"
//
// Note: While Brevo's API returns fields in camelCase (e.g., "firstName", "createdAt"),
// we normalize to lowercase for consistency with the field name handling across the platform.
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
// Brevo's standard objects are always plural: contacts, companies, deals, categories, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to lowercase.
// While Brevo's API uses camelCase (firstName, lastName), we normalize to lowercase
// for consistency with field name handling across the connector platform.
func normalizeFieldName(input string) string {
	return naming.ToLowerCase(input)
}

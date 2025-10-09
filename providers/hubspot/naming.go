package hubspot

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to HubSpot naming conventions.
// HubSpot uses lowercase plural for standard objects and lowercase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Contact" -> "contacts", "Company" -> "companies", "Deal" -> "deals"
//
// Fields:
//   - Converts to lowercase
//   - Examples: "FirstName" -> "firstname", "Email" -> "email"
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
// HubSpot's standard objects are always plural: contacts, companies, deals, tickets.
func normalizeObjectName(input string) string {
	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()
	// HubSpot uses lowercase
	return strings.ToLower(plural)
}

// normalizeFieldName converts field names to lowercase.
// HubSpot field names are case-insensitive but the API returns them in lowercase.
func normalizeFieldName(input string) string {
	return strings.ToLower(input)
}

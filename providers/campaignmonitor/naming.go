package campaignmonitor

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Campaign Monitor naming conventions.
// Campaign Monitor uses lowercase plural for object names in API endpoints and PascalCase for field names.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Client" -> "clients", "Admin" -> "admins", "Campaign" -> "campaigns"
//
// Fields:
//   - Converts to PascalCase (first letter of each word capitalized)
//   - Examples: "client_id" -> "ClientID", "email address" -> "EmailAddress", "name" -> "Name"
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
// Campaign Monitor's API endpoints use lowercase plural object names:
// /api/v3.3/clients.json, /api/v3.3/admins.json, /api/v3.3/campaigns.json.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to PascalCase.
// Campaign Monitor field names in API responses use PascalCase:
// ClientID, Name, EmailAddress, Status, FromName, FromEmail.
func normalizeFieldName(input string) string {
	return naming.CapitalizeFirstLetterEveryWord(input)
}

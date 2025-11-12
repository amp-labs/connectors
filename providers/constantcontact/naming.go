package constantcontact

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Constant Contact naming conventions.
// Constant Contact API v3 uses lowercase snake_case for both objects and fields, with plural
// object names.
//
// Objects:
//   - Converts to plural form with lowercase and underscores (snake_case)
//   - Examples: "Contact" -> "contacts", "ContactList" -> "contact_lists", "EmailCampaign" -> "email_campaigns"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "FirstName" -> "first_name", "EmailAddress" -> "email_address", "ContactId" -> "contact_id"
//
// Note: The API is case-sensitive and requires exact snake_case formatting for all entity names.
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

// normalizeObjectName converts object names to lowercase plural snake_case.
// Constant Contact objects are always plural: contacts, contact_lists, email_campaigns.
func normalizeObjectName(input string) string {
	// Convert to plural form and snake_case
	plural := naming.NewPluralString(input).String()

	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Constant Contact field names use snake_case: first_name, email_address, contact_id.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

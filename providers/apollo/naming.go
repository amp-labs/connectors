package apollo

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Apollo naming conventions.
// Apollo uses snake_case lowercase for objects and fields, with objects being plural.
//
// Objects:
//   - Converts to lowercase plural with underscores
//   - Examples: "Contact" -> "contacts", "Account" -> "accounts", "EmailerCampaign" -> "emailer_campaigns"
//   - Special mappings: "sequence" -> "emailer_campaigns", "deal" -> "opportunities"
//
// Fields:
//   - Converts to lowercase snake_case
//   - Examples: "FirstName" -> "first_name", "Email" -> "email", "CreatedAt" -> "created_at"
//
// API Evidence:
//   - Object names in responses: "contacts", "accounts", "opportunities", "emailer_campaigns"
//   - Field names: "first_name", "last_name", "created_at", "account_stage_id", "owner_id"
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
// Apollo's standard objects use plural form: contacts, accounts, opportunities, emailer_campaigns.
// The function also handles the product name mappings defined in objectNames.go.
func normalizeObjectName(input string) string {
	// First apply any display name to API name mappings
	// (e.g., "sequences" -> "emailer_campaigns", "deals" -> "opportunities")
	input = constructSupportedObjectName(input)

	// Convert to plural form and snake_case
	plural := naming.NewPluralString(input).String()

	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Apollo field names use snake_case: first_name, last_name, created_at, account_stage_id.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

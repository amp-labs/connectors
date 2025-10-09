package apollo

import (
	"context"
	"strings"

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

	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()

	// Convert to snake_case lowercase
	snakeCase := toSnakeCase(plural)

	return strings.ToLower(snakeCase)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Apollo field names use snake_case: first_name, last_name, created_at, account_stage_id.
func normalizeFieldName(input string) string {
	snakeCase := toSnakeCase(input)

	return strings.ToLower(snakeCase)
}

// toSnakeCase converts a string to snake_case.
// Handles transitions from lowercase to uppercase (e.g., "userId" -> "user_id").
// If the input already contains underscores, it's returned as-is.
//
//nolint:cyclop,mnd,nestif,varnamelen // Snake case conversion requires checking character patterns
func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	// If already in snake_case (contains underscore), return as-is
	if strings.Contains(input, "_") {
		return input
	}

	var result strings.Builder

	result.Grow(len(input) + 5) // Preallocate with buffer for underscores

	for idx, char := range input {
		// If this is an uppercase letter
		if char >= 'A' && char <= 'Z' {
			// Add underscore before uppercase if not the first character and previous was lowercase/digit
			if idx > 0 {
				prevChar := rune(input[idx-1])
				if (prevChar >= 'a' && prevChar <= 'z') || (prevChar >= '0' && prevChar <= '9') {
					result.WriteRune('_')
				}
			}
		}

		result.WriteRune(char)
	}

	return result.String()
}

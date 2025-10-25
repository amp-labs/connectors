package customerapp

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Customer.io naming conventions.
// Customer.io uses lowercase plural for objects with underscores and lowercase snake_case for fields.
//
// Objects:
//   - Converts to lowercase plural form with underscores preserved
//   - Examples: "Collection" -> "collections", "ReportingWebhook" -> "reporting_webhooks",
//     "SenderIdentity" -> "sender_identities", "SubscriptionTopic" -> "subscription_topics"
//
// Fields:
//   - Converts to lowercase with underscores preserved (snake_case)
//   - Examples: "CustomerId" -> "customer_id", "CreatedAt" -> "created_at",
//     "DeduplicateId" -> "deduplicate_id"
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
// Customer.io's standard objects are always lowercase plural: activities, broadcasts,
// collections, reporting_webhooks, sender_identities, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to lowercase.
// Customer.io field names are lowercase with underscores (snake_case).
// Examples: customer_id, created_at, deduplicate_id.
func normalizeFieldName(input string) string {
	return naming.ToLowerCase(input)
}

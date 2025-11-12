package blueshift

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Blueshift naming conventions.
// Blueshift uses lowercase plural for objects and lowercase snake_case for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Campaign" -> "campaigns", "Catalog" -> "catalogs", "EmailTemplate" -> "email_templates"
//
// Fields:
//   - Converts to lowercase (snake_case is already the standard)
//   - Examples: "Author" -> "author", "CreatedAt" -> "createdat", "UpdatedAt" -> "updatedat"
//
// Note: Blueshift API is consistent in using lowercase plural objects (campaigns, catalogs,
// email_templates, push_templates, sms_templates, etc.) and lowercase snake_case fields
// (author, created_at, updated_at, uuid, etc.).
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
// Blueshift's standard objects are always plural and lowercase:
// campaigns, catalogs, email_templates, push_templates, segments, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to lowercase.
// Blueshift field names are always lowercase with snake_case separators.
// Examples: author, created_at, updated_at, uuid, product_name_column_name.
func normalizeFieldName(input string) string {
	return naming.ToLowerCase(input)
}

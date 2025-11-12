package docusign

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to DocuSign naming conventions.
// DocuSign uses lowercase plural for objects (envelopes, templates) and camelCase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Envelope" -> "envelopes", "Template" -> "templates", "Document" -> "documents"
//
// Fields:
//   - Converts to camelCase (lowercase first letter, capitalize subsequent words)
//   - Examples: "EmailSubject" -> "emailSubject", "account_id" -> "accountId", "TEMPLATEID" -> "templateId"
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
// DocuSign's REST API uses lowercase plural for resource names in endpoints:
// /restapi/v2.1/accounts/{accountId}/envelopes, /templates, /documents, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to camelCase.
// DocuSign API uses camelCase for JSON field names (emailSubject, templateId, accountId, etc.)
func normalizeFieldName(input string) string {
	return naming.ToCamelCase(input)
}

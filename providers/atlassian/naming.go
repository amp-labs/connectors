package atlassian

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Atlassian (Jira) naming conventions.
// Jira uses lowercase singular for object names and lowercase for field names.
// Field names are case-insensitive but the API returns them in lowercase.
//
// Objects:
//   - Converts to lowercase singular form
//   - Examples: "Issue" -> "issue", "issues" -> "issue"
//
// Fields:
//   - Converts to lowercase
//   - Field names are case-insensitive in Jira API
//   - Custom fields use customfield_XXXXX format (e.g., "customfield_10000")
//   - Examples: "Summary" -> "summary", "IssueType" -> "issuetype", "CustomField_10000" -> "customfield_10000"
//
// Note: Currently only Jira is supported. If Confluence support is added in the future,
// this may need to handle different conventions (Confluence uses camelCase).
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
// Jira uses lowercase singular object names (e.g., "issue").
func normalizeObjectName(input string) string {
	// Convert to singular form using the naming package
	singular := naming.NewSingularString(input).String()
	// Jira uses lowercase
	return strings.ToLower(singular)
}

// normalizeFieldName converts field names to lowercase.
// Jira field names are case-insensitive but the API returns them in lowercase.
// Custom fields follow the pattern customfield_XXXXX.
func normalizeFieldName(input string) string {
	return strings.ToLower(input)
}

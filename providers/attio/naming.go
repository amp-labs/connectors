package attio

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Attio naming conventions.
// Attio uses snake_case for both objects and fields, with plural object names.
//
// Objects:
//   - Converts to lowercase snake_case plural form
//   - Examples: "Company" -> "companies", "WorkspaceMember" -> "workspace_members",
//     "List" -> "lists", "User" -> "users"
//
// Fields:
//   - Converts to lowercase snake_case
//   - Examples: "FirstName" -> "first_name", "EmailAddress" -> "email_address",
//     "ContentPlaintext" -> "content_plaintext"
//
// Attio API uses api_slug identifiers for objects and attributes. Standard objects include:
// - companies, people, deals, users (standard/custom objects)
// - lists, workspace_members, tasks, notes (special API objects)
//
// All field/attribute names use snake_case: name, email_address, first_name, last_name,
// record_id, user_id, created_at, content_plaintext, etc.
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

// normalizeObjectName converts object names to lowercase snake_case plural.
// Attio's standard objects are always plural: companies, people, deals, users,
// lists, workspace_members, tasks, notes.
func normalizeObjectName(input string) string {
	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()
	// Convert to snake_case (handles all case conversions)
	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Attio field names are always lowercase with underscores.
func normalizeFieldName(input string) string {
	// Convert to snake_case (handles all case conversions)
	return naming.ToSnakeCase(input)
}

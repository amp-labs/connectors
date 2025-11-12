package clickup

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to ClickUp naming conventions.
// ClickUp uses lowercase singular for objects and snake_case lowercase for fields.
//
// Objects:
//   - Converts to lowercase singular form
//   - Examples: "Team" -> "team", "Tasks" -> "task", "List" -> "list"
//
// Fields:
//   - Converts to snake_case lowercase
//   - Examples: "ListId" -> "list_id", "TimeSpent" -> "time_spent", "firstName" -> "first_name"
//
// Note: ClickUp API responses use plural keys (e.g., "teams") but the URL paths and
// object names in the API are singular (e.g., GET /api/v2/team).
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
// ClickUp's API paths use singular object names: /team, /task, /list, /space.
func normalizeObjectName(input string) string {
	// Convert to singular form and lowercase
	singular := naming.NewSingularString(input).String()

	return naming.ToLowerCase(singular)
}

// normalizeFieldName converts field names to snake_case lowercase.
// ClickUp field names use snake_case: list_id, time_spent, home_list.
func normalizeFieldName(input string) string {
	// Convert to snake_case (handles PascalCase, camelCase, etc.)
	snakeCase := naming.ToSnakeCase(input)

	// Ensure lowercase
	return naming.ToLowerCase(snakeCase)
}

package asana

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Asana naming conventions.
// Asana uses snake_case plural for object names (resources) and snake_case for field names.
//
// Objects:
//   - Converts to lowercase plural with underscores
//   - Examples: "Task" -> "tasks", "Project" -> "projects", "CustomField" -> "custom_fields"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "CreatedAt" -> "created_at", "ResourceType" -> "resource_type", "gid" -> "gid"
//
// The Asana API documentation shows consistent use of snake_case for both resource names
// and field names. Resource endpoints are plural (e.g., /tasks, /projects, /custom_fields).
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

// normalizeObjectName converts object names to snake_case plural.
// Asana's API uses plural resource names: tasks, projects, workspaces, custom_fields, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and snake_case
	plural := naming.NewPluralString(input).String()

	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to snake_case.
// Asana field names consistently use snake_case: gid, resource_type, created_at, custom_fields.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

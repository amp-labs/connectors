package asana

import (
	"context"
	"strings"

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
	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()
	// Convert to snake_case and lowercase
	return toSnakeCase(strings.ToLower(plural))
}

// normalizeFieldName converts field names to snake_case.
// Asana field names consistently use snake_case: gid, resource_type, created_at, custom_fields.
func normalizeFieldName(input string) string {
	return toSnakeCase(strings.ToLower(input))
}

// toSnakeCase converts a string to snake_case.
// It handles strings that may be in PascalCase, camelCase, or already have underscores.
func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	// If it already contains underscores and is lowercase, return as-is
	if strings.Contains(input, "_") && input == strings.ToLower(input) {
		return input
	}

	var result strings.Builder

	const extraCapacity = 5

	result.Grow(len(input) + extraCapacity) // Pre-allocate some extra space for underscores

	const asciiUpperToLower = 32

	for i, char := range input {
		// If we hit an uppercase letter (shouldn't happen after ToLower, but keep for safety)
		if char >= 'A' && char <= 'Z' {
			// Add underscore before uppercase if not at start
			if i > 0 {
				result.WriteRune('_')
			}

			result.WriteRune(char + asciiUpperToLower) // Convert to lowercase
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

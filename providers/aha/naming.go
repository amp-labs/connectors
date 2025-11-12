package aha

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Aha API naming conventions.
// Aha uses lowercase plural with underscores for objects in API paths (e.g., /api/v1/features),
// and lowercase with underscores (snake_case) for field names.
//
// Objects:
//   - Converts to plural lowercase with underscores for multi-word names
//   - Examples: "feature" -> "features", "Feature" -> "features", "IdeaPortal" -> "idea_portals"
//   - Special case: compound paths like "ideas/endorsements" remain unchanged
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "CreatedAt" -> "created_at", "WorkflowStatus" -> "workflow_status"
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

// normalizeObjectName converts object names to plural lowercase with underscores.
// Aha's standard objects are plural: features, products, ideas, releases, goals.
// Multi-word objects use underscores: idea_portals, team_members, release_phases.
// Compound paths with slashes (e.g., "ideas/endorsements") are preserved.
func normalizeObjectName(input string) string {
	// If the input contains a slash, it's a compound path (e.g., "ideas/endorsements")
	// These should be handled specially and not pluralized again
	if strings.Contains(input, "/") {
		// Split by slash, normalize each part, and rejoin
		parts := strings.Split(input, "/")
		for i, part := range parts {
			parts[i] = normalizeSimpleObjectName(part)
		}

		return strings.Join(parts, "/")
	}

	return normalizeSimpleObjectName(input)
}

// normalizeSimpleObjectName handles normalization for a single object name (no slashes).
func normalizeSimpleObjectName(input string) string {
	// Convert to plural form and snake_case
	plural := naming.NewPluralString(input).String()

	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase with underscores (snake_case).
// Aha field names are case-insensitive but the API returns them in snake_case.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

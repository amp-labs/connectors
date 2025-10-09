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
	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()

	// Convert to lowercase with underscores (snake_case)
	return toSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase with underscores (snake_case).
// Aha field names are case-insensitive but the API returns them in snake_case.
func normalizeFieldName(input string) string {
	return toSnakeCase(input)
}

// toSnakeCase converts a string to snake_case.
// Handles PascalCase, camelCase, and existing snake_case inputs.
//
//nolint:cyclop // Snake case conversion requires multiple conditions
func toSnakeCase(s string) string {
	// Simple cases: return early
	lower := strings.ToLower(s)
	if s == "" || s == lower {
		return lower
	}

	var result strings.Builder

	runes := []rune(s)

	for idx, char := range runes {
		// Add underscore before uppercase letters (except first character)
		if idx > 0 && isUpper(char) && shouldAddUnderscore(runes, idx) {
			result.WriteRune('_')
		}

		result.WriteRune(toLower(char))
	}

	return result.String()
}

// shouldAddUnderscore determines if an underscore should be added before the current character.
func shouldAddUnderscore(runes []rune, idx int) bool {
	// Don't add if previous char is already underscore
	if runes[idx-1] == '_' {
		return false
	}

	// Add if previous char is lowercase (transition from lower to upper)
	if isLower(runes[idx-1]) {
		return true
	}

	// For uppercase sequences, add underscore before the last uppercase
	// if it's followed by lowercase (e.g., "HTTPResponse" -> "http_response")
	if idx+1 < len(runes) && isLower(runes[idx+1]) {
		return true
	}

	return false
}

// isUpper checks if a rune is an uppercase letter.
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// isLower checks if a rune is a lowercase letter.
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// toLower converts a rune to lowercase.
func toLower(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}

	return r
}

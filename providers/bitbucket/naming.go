package bitbucket

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Bitbucket naming conventions.
// Bitbucket uses lowercase with underscores (snake_case) for both objects and fields.
//
// Objects:
//   - Converts to lowercase plural form with underscores
//   - Examples: "Repository" -> "repositories", "PullRequest" -> "pull_requests"
//   - Special handling: multi-word objects use underscores (e.g., "branch_restrictions")
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "CreatedOn" -> "created_on", "FullName" -> "full_name"
//   - Common fields: "updated_on", "created_on", "full_name", "display_name"
//
// Bitbucket API Reference:
//   - Base URL: https://api.bitbucket.org/2.0/
//   - Objects: repositories, workspaces, pullrequests, pipelines, etc.
//   - Fields use snake_case: created_on, updated_on, full_name, etc.
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

// normalizeObjectName converts object names to lowercase plural with underscores.
// Bitbucket's standard objects are always plural and use underscores:
// repositories, workspaces, pullrequests, branch_restrictions, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form using the naming package
	plural := naming.NewPluralString(input).String()

	// Convert to snake_case (lowercase with underscores)
	return toSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase with underscores (snake_case).
// Bitbucket field names consistently use snake_case:
// created_on, updated_on, full_name, display_name, etc.
func normalizeFieldName(input string) string {
	return toSnakeCase(input)
}

// toSnakeCase converts a string to snake_case format.
// Examples:
//   - "CreatedOn" -> "created_on"
//   - "FullName" -> "full_name"
//   - "PullRequest" -> "pull_request"
//   - "branch_restrictions" -> "branch_restrictions" (already snake_case)
func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	// Handle already lowercase strings with underscores
	if isSnakeCase(input) {
		return input
	}

	const extraCapacity = 5 // Preallocate space for potential underscores

	var result strings.Builder

	result.Grow(len(input) + extraCapacity)

	for i, r := range input {
		if i > 0 && isUpperCase(r) {
			// Add underscore before uppercase letters (except at start)
			// But don't add if previous char was already an underscore
			if i > 0 && input[i-1] != '_' {
				result.WriteRune('_')
			}
		}

		result.WriteRune(toLowerRune(r))
	}

	return result.String()
}

// isSnakeCase checks if a string is already in snake_case format.
func isSnakeCase(input string) bool {
	for _, r := range input {
		if isUpperCase(r) {
			return false
		}
	}

	return true
}

// isUpperCase checks if a rune is an uppercase letter.
func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// toLowerRune converts a rune to lowercase.
func toLowerRune(r rune) rune {
	const asciiOffset = 32 // Offset to convert uppercase to lowercase in ASCII

	if isUpperCase(r) {
		return r + asciiOffset
	}

	return r
}

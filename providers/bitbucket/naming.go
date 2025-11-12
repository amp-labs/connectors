package bitbucket

import (
	"context"

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
	// Convert to plural form and snake_case
	plural := naming.NewPluralString(input).String()

	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase with underscores (snake_case).
// Bitbucket field names consistently use snake_case:
// created_on, updated_on, full_name, display_name, etc.
func normalizeFieldName(input string) string {
	return naming.ToSnakeCase(input)
}

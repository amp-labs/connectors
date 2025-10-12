package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors"
)

// NormalizeEntityName normalizes entity names according to Chili Piper naming conventions.
// Chili Piper uses lowercase singular objects with optional path segments (e.g., "workspace",
// "team", "workspace/users") and camelCase fields. The API is case-sensitive and expects
// exact names, so this implementation uses a pass-through approach.
//
// Objects:
//   - Already lowercase singular or path-based (workspace, team, distribution, workspace/users)
//   - No transformation needed - return as-is
//   - Examples: "workspace" -> "workspace", "team" -> "team", "workspace/users" -> "workspace/users"
//
// Fields:
//   - Already camelCase (id, name, workspaceId, createdAt, teamMembersMetadata)
//   - No transformation needed - return as-is
//   - Examples: "id" -> "id", "workspaceId" -> "workspaceId", "createdAt" -> "createdAt"
//
// Note: The Chili Piper Edge API uses consistent naming conventions across all endpoints,
// eliminating the need for normalization. Both objects and fields should be used exactly
// as they appear in the API documentation.
func (c *Connector) NormalizeEntityName(
	ctx context.Context, entity connectors.Entity, input string,
) (normalized string, err error) {
	// Chili Piper API uses consistent naming - no normalization required
	// Return input unchanged for all entity types
	return input, nil
}

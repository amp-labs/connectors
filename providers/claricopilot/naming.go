package claricopilot

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Clari Copilot naming conventions.
// Clari Copilot uses lowercase plural for most objects (with some exceptions) and mixed case for fields.
//
// Objects:
//   - Converts to lowercase plural form for most objects
//   - Special cases: "scorecard" remains singular, "scorecard-template" uses hyphenation
//   - Examples: "Call" -> "calls", "User" -> "users", "Contact" -> "contacts", "Deal" -> "deals"
//   - Note: Write operations internally convert to singular (handled by writeObjectMapping)
//
// Fields:
//   - Returns unchanged (preserves original case)
//   - Clari Copilot uses mixed conventions: snake_case (source_id, last_modified_time)
//     and camelCase (userId, isOrganizer) depending on the field
//   - The API is case-sensitive and expects exact field names as returned
//   - Examples: "source_id" -> "source_id", "userId" -> "userId", "isOrganizer" -> "isOrganizer"
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

// normalizeObjectName converts object names to lowercase plural.
// Clari Copilot's API endpoints use lowercase plural for most objects.
// Special cases like "scorecard" and "scorecard-template" are preserved.
func normalizeObjectName(input string) string {
	// Handle special cases that should remain singular or have special formatting
	switch naming.ToLowerCase(input) {
	case objectNameScorecard, "scorecards":
		return objectNameScorecard
	case objectNameScorecardTemplate, "scorecard-templates":
		return objectNameScorecardTemplate
	}

	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName returns the field name unchanged.
// Clari Copilot uses mixed field naming conventions (both snake_case and camelCase)
// and the API is case-sensitive, so we preserve the exact field names as provided.
func normalizeFieldName(input string) string {
	return input
}

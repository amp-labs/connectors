package braze

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Braze naming conventions.
// Braze uses lowercase for all objects and fields, but has inconsistent pluralization
// and special characters (underscores, slashes) in object names.
//
// Objects:
//   - Converts to lowercase
//   - Preserves plurality as-is (Braze uses mixed: "campaigns" is plural, "canvas" is singular)
//   - Preserves special characters (underscores and slashes)
//   - Examples: "Campaigns" -> "campaigns", "Canvas" -> "canvas", "PreferenceCenter" -> "preferencecenter"
//
// Note: Due to Braze's inconsistent naming patterns (mixed plurality, underscores, slashes),
// we use a simple lowercase normalization strategy rather than attempting to normalize plurality.
// Object names like "preference_center", "content_blocks", and "templates/email" are preserved as-is
// after lowercase conversion.
//
// Fields:
//   - Converts to lowercase
//   - Examples: "FirstName" -> "firstname", "Email" -> "email"
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

// normalizeObjectName converts object names to lowercase.
// Braze object names are case-insensitive but the API uses lowercase consistently.
// We preserve the input structure (including special characters like underscores and slashes)
// since Braze uses mixed patterns: "campaigns" (plural), "canvas" (singular),
// "preference_center" (underscore), "templates/email" (slash).
func normalizeObjectName(input string) string {
	return naming.ToLowerCase(input)
}

// normalizeFieldName converts field names to lowercase.
// Braze field names follow the same lowercase convention as objects.
func normalizeFieldName(input string) string {
	return naming.ToLowerCase(input)
}

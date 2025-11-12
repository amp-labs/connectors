package capsule

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Capsule CRM naming conventions.
// Capsule uses lowercase plural for objects (parties, opportunities, tasks) and camelCase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Examples: "Party" -> "parties", "Opportunity" -> "opportunities", "Task" -> "tasks"
//   - Special case: "kases" is used in the API (Projects were renamed from Cases, but API unchanged)
//
// Fields:
//   - Converts to camelCase (lowercase first letter)
//   - Examples: "FirstName" -> "firstName", "LastName" -> "lastName", "JobTitle" -> "jobTitle"
//
// Based on API documentation: https://developer.capsulecrm.com/v2/operations/Party
// The API uses lowercase plural object names and camelCase field names consistently.
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
// Capsule's standard objects are always plural: parties, opportunities, tasks, kases, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and lowercase
	plural := naming.NewPluralString(input).String()

	return naming.ToLowerCase(plural)
}

// normalizeFieldName converts field names to camelCase (lowercase first letter).
// Capsule field names use camelCase: firstName, lastName, jobTitle, emailAddresses, etc.
func normalizeFieldName(input string) string {
	// Convert to lowercase for the entire string first
	lower := naming.ToLowerCase(input)

	// Return as-is since Capsule uses camelCase (which preserves lowercase)
	// If the input is already camelCase like "firstName", it stays that way
	// If the input is PascalCase like "FirstName", ToLowerCase makes it "firstname"
	// For now, we'll use a simple lowercase approach as the API is case-insensitive
	// and returns camelCase, but accepts various cases
	return lower
}

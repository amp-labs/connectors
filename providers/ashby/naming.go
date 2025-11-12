package ashby

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Ashby naming conventions.
// Ashby uses camelCase for both objects and fields, with objects in singular form.
//
// Objects:
//   - Converts to singular camelCase (first letter lowercase)
//   - Examples: "Application" -> "application", "Candidates" -> "candidate"
//   - Compound names: "InterviewSchedule" -> "interviewSchedule", "job_posting" -> "jobPosting"
//
// Fields:
//   - Converts to camelCase (first letter lowercase)
//   - Examples: "FirstName" -> "firstName", "is_archived" -> "isArchived"
//
// Note: Ashby API is case-sensitive and requires exact camelCase matching.
// The API uses dot notation for operations (e.g., "candidate.list", "application.create").
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

// normalizeObjectName converts object names to singular camelCase.
// Ashby's standard objects are singular and camelCase: application, candidate, job, user.
func normalizeObjectName(input string) string {
	// Convert to singular form and camelCase
	singular := naming.NewSingularString(input).String()

	return naming.ToCamelCase(singular)
}

// normalizeFieldName converts field names to camelCase.
// Ashby field names use camelCase: firstName, isArchived, primaryEmailAddress.
func normalizeFieldName(input string) string {
	return naming.ToCamelCase(input)
}

package aws

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to AWS API naming conventions.
// AWS APIs use PascalCase for both object names and field names, with plural object names.
//
// Objects:
//   - Converts to PascalCase with first letter capitalized
//   - Uses plural form (e.g., "Users", "Groups", "Applications")
//   - Examples: "user" -> "Users", "group" -> "Groups", "application" -> "Applications"
//
// Fields:
//   - Converts to PascalCase with first letter capitalized
//   - Examples: "userId" -> "UserId", "displayname" -> "DisplayName", "email" -> "Email"
//
// This applies to AWS Identity Center (formerly AWS SSO) and other AWS services
// which consistently use PascalCase naming in their JSON APIs.
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

// normalizeObjectName converts object names to PascalCase plural.
// AWS objects are plural and use PascalCase: Users, Groups, Applications, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form and PascalCase
	plural := naming.NewPluralString(input).String()

	return naming.ToPascalCase(plural)
}

// normalizeFieldName converts field names to PascalCase.
// AWS field names use PascalCase: UserId, DisplayName, ApplicationArn, etc.
func normalizeFieldName(input string) string {
	return naming.ToPascalCase(input)
}

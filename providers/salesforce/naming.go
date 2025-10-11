package salesforce

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Salesforce naming conventions.
// Salesforce uses PascalCase for standard objects (Account, Contact) and requires exact
// case matching for API calls. Custom objects use __c suffix (MyObject__c).
//
// Objects:
//   - Converts to PascalCase with first letter capitalized
//   - Preserves __c suffix for custom objects
//   - Examples: "account" -> "Account", "contact" -> "Contact", "myobject__c" -> "Myobject__c"
//
// Fields:
//   - Converts to lowercase (Salesforce's internal representation)
//   - Preserves __c suffix for custom fields
//   - Examples: "FirstName" -> "firstname", "Email" -> "email", "Custom__c" -> "custom__c"
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

// normalizeObjectName converts object names to PascalCase.
// Standard Salesforce objects are PascalCase (Account, Contact, Lead).
// Custom objects end with __c and the base name should be capitalized.
func normalizeObjectName(input string) string {
	// Check if it's a custom object
	if strings.HasSuffix(input, "__c") {
		// Extract base name without __c suffix
		base := strings.TrimSuffix(input, "__c")
		// Capitalize first letter and append __c
		return capitalizeFirst(base) + "__c"
	}

	// For standard objects, use singular form and capitalize
	singular := naming.NewSingularString(input).String()

	return capitalizeFirst(singular)
}

// normalizeFieldName converts field names to lowercase.
// Salesforce's internal metadata representation uses lowercase field names.
// Custom fields end with __c and should preserve that suffix.
func normalizeFieldName(input string) string {
	return strings.ToLower(input)
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

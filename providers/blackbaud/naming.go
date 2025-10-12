package blackbaud

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Blackbaud SKY API naming conventions.
// Blackbaud uses lowercase plural for objects and lowercase for fields.
//
// Objects:
//   - Converts to lowercase plural form
//   - Note: Object names in Blackbaud often include module prefixes (e.g., "crm-adnmg/batchtemplates")
//     but this normalization applies to the object name portion only
//   - Examples: "BatchTemplate" -> "batchtemplates", "Currency" -> "currencies", "Event" -> "events"
//   - Examples with prefix: "crm-adnmg/BatchTemplate" -> "crm-adnmg/batchtemplates"
//
// Fields:
//   - Converts to lowercase
//   - Blackbaud field names are returned in lowercase from the API
//   - Examples: "FirstName" -> "firstname", "Email" -> "email", "ConstituentId" -> "constituentid"
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
// Blackbaud SKY API uses lowercase plural for object names across all modules.
// Examples from CRM:
//   - batchtemplates, currencies, sites (CRM Administration)
//   - events, locations, registrants (CRM Event)
//   - payments, revenuetransactions (CRM Revenue)
//   - volunteers, occurrences (CRM Volunteer)
//
// If the input contains a module prefix (e.g., "crm-adnmg/BatchTemplate"),
// only the object portion after the last "/" is normalized.
func normalizeObjectName(input string) string {
	// Handle module prefixes (e.g., "crm-adnmg/batchtemplates")
	// Split on last "/" to separate module from object name
	lastSlash := strings.LastIndex(input, "/")

	var prefix string

	var objectName string

	if lastSlash >= 0 {
		prefix = input[:lastSlash+1] // Include the "/"
		objectName = input[lastSlash+1:]
	} else {
		objectName = input
	}

	// Convert to plural form using the naming package
	plural := naming.NewPluralString(objectName).String()

	// Blackbaud uses lowercase
	normalizedObject := strings.ToLower(plural)

	return prefix + normalizedObject
}

// normalizeFieldName converts field names to lowercase.
// Blackbaud field names are case-insensitive but the API returns them in lowercase.
func normalizeFieldName(input string) string {
	return strings.ToLower(input)
}

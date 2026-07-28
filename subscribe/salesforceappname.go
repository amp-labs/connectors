package subscribe

import "strings"

// SanitizeAppNameForSalesforce normalizes an app name for use in Salesforce custom field names.
// It lowercases all characters, keeps alphanumerics and underscores, replaces hyphens
// and spaces with underscores, removes all other characters, and trims trailing underscores.
//
// Mirrors the server's salesforce.SanitizeAppNameForSalesforce.
func SanitizeAppNameForSalesforce(appName string) string {
	var result strings.Builder

	for _, char := range appName {
		switch {
		case char >= 'a' && char <= 'z',
			char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9',
			char == '_':
			result.WriteRune(char)
		case char == '-' || char == ' ':
			result.WriteRune('_')
		}
	}

	return strings.TrimRight(strings.ToLower(result.String()), "_")
}

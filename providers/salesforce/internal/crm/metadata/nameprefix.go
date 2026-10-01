package metadata

import (
	"strings"
)

// developerNameMaxLength is Salesforce's cap on a Flow API name and an outbound
// message developer name.
const developerNameMaxLength = 80

// DefaultNamePrefix names artifacts when the caller supplies no usable prefix.
const DefaultNamePrefix = "amp"

// letterPrefix is prepended when the sanitized project name does not start with
// a letter (Salesforce developer names must). That keeps digit-leading names
// like "11x" intact as "proj_11x" instead of stripping them down to "x".
const letterPrefix = "proj"

// SanitizeNamePrefix normalizes a caller-supplied prefix (typically the builder's
// project or app name) into something Salesforce accepts at the front of a
// developer name: letters, digits and single underscores, beginning with a letter.
//
//nolint:cyclop
func SanitizeNamePrefix(prefix string) string {
	var cleaned strings.Builder

	for _, char := range prefix {
		switch {
		case char >= 'a' && char <= 'z',
			char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9',
			char == '_':
			cleaned.WriteRune(char)
		case char == '-' || char == ' ' || char == '.':
			cleaned.WriteRune('_')
		}
	}

	result := cleaned.String()
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	result = strings.Trim(result, "_")
	if result == "" {
		return ""
	}

	if !isASCIILetter(rune(result[0])) {
		return letterPrefix + "_" + result
	}

	return result
}

func isASCIILetter(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

// Suffixes for the per-event artifact names of an object's create and update flow pairs.
const (
	ArtifactSuffixCreate = "Create"
	ArtifactSuffixUpdate = "Update"
)

// GenerateSubscriptionArtifactName builds the developer name shared by one
// event type's flow and its outbound message: "<prefix>_<Object>_<suffix>"
// (e.g. "acme_Account_Create", or "acme_My_Object_c_Update" — custom-object
// suffixes like "__c" collapse, since developer names cannot contain consecutive
// underscores). An empty suffix gives "<prefix>_<Object>". The prefix is
// truncated to keep the whole name within developerNameMaxLength.
func GenerateSubscriptionArtifactName(prefix, objectName, suffix string) (string, error) {
	if objectName == "" {
		return "", errEmptyObjectName
	}

	cleanPrefix := SanitizeNamePrefix(prefix)
	if cleanPrefix == "" {
		cleanPrefix = DefaultNamePrefix
	}

	cleanObject := objectName
	for strings.Contains(cleanObject, "__") {
		cleanObject = strings.ReplaceAll(cleanObject, "__", "_")
	}

	tail := cleanObject
	if suffix != "" {
		tail += "_" + suffix
	}

	// One underscore joins the prefix to the rest.
	available := developerNameMaxLength - len(tail) - 1
	if len(cleanPrefix) > available {
		cleanPrefix = strings.TrimRight(cleanPrefix[:available], "_")
	}

	return cleanPrefix + "_" + tail, nil
}

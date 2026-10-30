package httpkit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type PathIdentifier struct {
	Name  string
	Value string
}

// ResolveUrlPathTemplate replaces path identifiers in a URL path template with their corresponding values.
//
// Identifiers are specified using braces, for example, {record_id}. Each
// identifier in the template must have a corresponding PathIdentifier.
//
// If the template contains no identifiers, it is returned unchanged.
//
// Returns an error if a required identifier is missing or has an empty value.
// Missing identifiers indicate an invalid implementation, while empty values
// indicate that the called could not supply a required value.
func ResolveUrlPathTemplate(pathTemplate string, ids ...PathIdentifier) (string, error) {
	requiredIdentifiers := ExtractPathIds(pathTemplate)
	if len(requiredIdentifiers) == 0 {
		return pathTemplate, nil
	}

	registry := make(map[string]PathIdentifier, len(ids))
	for _, id := range ids {
		registry[id.Name] = id
	}

	result := pathTemplate

	for _, name := range requiredIdentifiers {
		identifier, ok := registry[name]
		if !ok {
			return "", fmt.Errorf(
				"%w: missing required path identifier %q",
				common.ErrInvalidImplementation, name,
			)
		}

		if identifier.Value == "" {
			return "", fmt.Errorf("%w: empty value for path identifier %q: connector missing value to supply it",
				common.ErrResolvingURLPathForObject, name)
		}

		result = strings.ReplaceAll(result, "{"+name+"}", identifier.Value)
	}

	return result, nil
}

// ExtractPathIds returns the names of identifiers enclosed in braces in a URL path.
//
// Identifiers are returned in the order they appear in the path.
// Example: /v3/orders/{order_id}/line_items/{line_items_id}
func ExtractPathIds(urlPath string) []string {
	varIDPattern := regexp.MustCompile(`\{([^{}]+)\}`)

	matches := varIDPattern.FindAllStringSubmatch(urlPath, -1)

	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}

	return ids
}

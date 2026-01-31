package readhelper

import (
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// SelectedFieldsFunc returns a whitelist of the original record, it handles the nested fields too.
// Second output is the identifier of this record.
type SelectedFieldsFunc func(node *ajson.Node) (map[string]any, string, error)

func MakeMarshaledSelectedDataFunc(
	selectFields SelectedFieldsFunc,
	rawTransformer common.RecordTransformer,
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		if selectFields == nil {
			selectFields = func(node *ajson.Node) (map[string]any, string, error) {
				allFields, err := jsonquery.Convertor.ObjectToMap(node)
				if err != nil {
					return nil, "", err
				}

				return allFields, "", nil
			}
		}

		if rawTransformer == nil {
			rawTransformer = func(node *ajson.Node) (map[string]any, error) {
				return jsonquery.Convertor.ObjectToMap(node)
			}
		}

		data := make([]common.ReadResultRow, len(records))

		for index, nodeRecord := range records {
			raw, err := rawTransformer(nodeRecord)
			if err != nil {
				return nil, err
			}

			selectedFields, identifier, err := selectFields(nodeRecord)
			if err != nil {
				return nil, err
			}

			data[index] = common.ReadResultRow{
				Fields: selectedFields,
				Id:     identifier,
				Raw:    raw,
			}
		}

		return data, nil
	}
}

// SelectFields returns a new map containing only the whitelisted fields.
func SelectFields(
	record map[string]any,
	fields datautils.StringSet,
) map[string]any {
	output := make(map[string]any)

	for field := range fields {
		path := parseJSONPath(field)
		if value, ok := getNestedValue(record, path); ok {
			setNestedValue(output, path, value)
		}
	}

	return output
}

// parseJSONPath converts a JSONPath-like string into a slice of keys.
// Example:
//
//	$['payload']['body']['data'] -> ["payload", "body", "data"]
//
// If the input does not match the $[...] pattern, it is returned as a single key.
func parseJSONPath(path string) []string {
	// regex to match ['key']
	re := regexp.MustCompile(`\['([^']+)'\]`)
	matches := re.FindAllStringSubmatch(path, -1)

	// if no matches, return the path itself as a single key
	if len(matches) == 0 {
		return []string{path}
	}

	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			keys = append(keys, m[1])
		}
	}

	return keys
}

// getNestedValue retrieves a value from a nested map following the path.
// Returns false if any key along the path is missing or not a map (except the leaf).
func getNestedValue(m map[string]any, path []string) (any, bool) {
	curr := m
	for i, key := range path {
		v, ok := curr[key]
		if !ok {
			return nil, false
		}

		if i == len(path)-1 {
			return v, true
		}

		nextMap, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}

		curr = nextMap
	}

	return nil, false
}

// setNestedValue inserts a value into a map at the specified path.
// Creates intermediate maps if needed.
func setNestedValue(m map[string]any, path []string, value any) {
	curr := m

	for i, rawKey := range path {
		// The key must be case-insensitive.
		// The record returned by ReadConnector has keys always in lower case.
		key := strings.ToLower(rawKey)

		if i == len(path)-1 {
			curr[key] = value

			return
		}

		next, ok := curr[key]
		if !ok {
			newMap := make(map[string]any)
			curr[key] = newMap
			curr = newMap

			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			newMap := make(map[string]any)
			curr[key] = newMap
			curr = newMap

			continue
		}

		curr = nextMap
	}
}

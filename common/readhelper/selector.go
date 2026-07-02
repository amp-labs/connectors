package readhelper

import (
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// ArrayWildIndex represents a wildcard index in JSONPath notation.
//
// Normally, a concrete numeric index is used to access a specific array item.
// For our connectors, we often need to match all items in an array, so `*` is used instead.
//
// Example:
//
//	"$['line_items']['data'][*]['description']"
//	==> line_items[data][0,1,2, ... ][description]
const ArrayWildIndex = "*"

// SelectedFieldsFunc returns a whitelist of the original record, it handles the nested fields too.
// Second output is the identifier of this record.
type SelectedFieldsFunc func(node *ajson.Node, fields []string) (map[string]any, string, error)

func MakeMarshaledSelectedDataFunc(
	selectFields SelectedFieldsFunc,
	rawTransformer common.RecordTransformer,
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		if selectFields == nil {
			selectFields = func(node *ajson.Node, f []string) (map[string]any, string, error) {
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

			selectedFields, identifier, err := selectFields(nodeRecord, fields)
			if err != nil {
				return nil, err
			}

			row := common.ReadResultRow{
				Id:  identifier,
				Raw: raw,
			}

			if len(selectedFields) != 0 {
				row.Fields = selectedFields
			}

			data[index] = row
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
		path := ParseJSONPath(field)
		if value, ok := getNestedValue(record, path); ok {
			setNestedValue(output, path, value)
		}
	}

	return output
}

// ParseJSONPath splits a JSONPath-like expression into path tokens.
//
// Supported tokens are object keys in bracket form and array wildcards:
//
//	$['payload']['body']['data'] -> ["payload", "body", "data"]
//	$['items'][*]['currency']     -> ["items", "*", "currency"]
//
// If the input does not contain any supported path tokens, the original
// string is returned as a single-element path.
func ParseJSONPath(path string) []string {
	// regex to match ['key'] or [*]
	re := regexp.MustCompile(`\['([^']+)'\]|\[(\*)\]`)
	matches := re.FindAllStringSubmatch(path, -1)

	// if no matches, return the path itself as a single key
	if len(matches) == 0 {
		return []string{path}
	}

	// A match is an array of tuples [][]string.
	// Here is an example of some tuples:
	//	[0]: "['line_items']"
	//	[1]: "line_items"
	//	[2]: ""
	//
	//	[0]: "[*]"
	//	[1]: ""
	//	[2]: "*"
	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 && m[1] != "" {
			keys = append(keys, m[1])
		} else if len(m) > 2 && m[2] == ArrayWildIndex {
			keys = append(keys, ArrayWildIndex)
		}
	}

	return keys
}

// getNestedValue retrieves a value from a nested structure using the supplied path.
//
// The input may contain nested maps and arrays. A path token of "*" is treated as
// an array wildcard, meaning the next path segment is applied to every element
// of the array.
//
// Examples:
//
//	["payload", "body", "data"]       -> map traversal only
//	["items", "*", "currency"]        -> collect currency from each array item
//
// The function returns false if any path segment is missing, has the wrong type,
// or if the path cannot be resolved.
func getNestedValue(root any, path []string) (any, bool) { // nolint:lll,cyclop
	if len(path) == 0 {
		return root, true
	}

	switch node := root.(type) {
	case map[string]any:
		key := path[0]

		child, ok := node[key]
		if !ok {
			return nil, false
		}

		// If the next token is "*", the current value must be an array.
		if len(path) > 1 && path[1] == ArrayWildIndex {
			arr, ok := child.([]any)
			if !ok {
				return nil, false
			}

			if len(path) < 3 { // nolint:mnd
				return nil, false
			}

			result := make([]any, 0, len(arr))
			for _, item := range arr {
				v, ok := getNestedValue(item, path[2:])
				if !ok {
					return nil, false
				}

				result = append(result, v)
			}

			return result, true
		}

		return getNestedValue(child, path[1:])

	case []any:
		return nil, false

	default:
		return node, len(path) == 0
	}
}

// setNestedValue inserts a value into a nested map at the specified path.
//
// Intermediate maps and arrays are created as needed. Array paths use "*"
// as the wildcard token, and the value must be a []any when writing through
// an array segment.
func setNestedValue(root map[string]any, path []string, value any) {
	setNode(root, path, value)
}

func setNode(current any, path []string, value any) any { // nolint:cyclop
	if len(path) == 0 {
		return value
	}

	rawKey := path[0]

	switch node := current.(type) {
	case map[string]any:
		// Keys are normalized to lowercase because connector payload keys are lowercased.
		key := strings.ToLower(rawKey)
		if len(path) == 1 {
			node[key] = value

			return current
		}

		next, ok := node[key]
		if !ok || next == nil {
			next = makeContainer(path[1])
			node[key] = next
		}

		node[key] = setNode(next, path[1:], value)

		return current

	case []any:
		vals, ok := value.([]any)
		if !ok {
			return current
		}

		if len(node) < len(vals) {
			tmp := make([]any, len(vals))
			copy(tmp, node)
			node = tmp
		}

		for i := range vals {
			if node[i] == nil {
				node[i] = make(map[string]any)
			}

			node[i] = setNode(node[i], path[1:], vals[i])
		}

		return node
	}

	return current
}

func makeContainer(next string) any {
	switch next {
	case ArrayWildIndex:
		return []any{}
	default:
		return make(map[string]any)
	}
}

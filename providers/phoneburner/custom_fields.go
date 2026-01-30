package phoneburner

import (
	"strings"
	"unicode"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// phoneburnerCustomFieldKey converts a human label into a stable field key
// that can be requested via Read and returned by ListObjectMetadata.
//
// Example: "My custom field" -> "my_custom_field".
func phoneburnerCustomFieldKey(label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	if label == "" {
		return ""
	}

	// Replace non [a-z0-9] with underscores, and collapse repeats.
	var b strings.Builder
	b.Grow(len(label))

	lastUnderscore := false
	for _, r := range label {
		isAlnum := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isAlnum {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}

		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return ""
	}

	// Field keys shouldn't start with a digit.
	if out[0] >= '0' && out[0] <= '9' {
		out = "custom_" + out
	}

	return out
}

// flattenContactCustomFields promotes values from the "custom_fields" array into root-level keys.
func flattenContactCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFields, err := jsonquery.New(node).ArrayOptional("custom_fields")
	if err != nil || len(customFields) == 0 {
		return root, err
	}

	for _, cfNode := range customFields {
		q := jsonquery.New(cfNode)

		name, err := q.TextWithDefault("name", "")
		if err != nil {
			return nil, err
		}
		if name == "" {
			name, err = q.TextWithDefault("display_name", "")
			if err != nil {
				return nil, err
			}
		}

		key := phoneburnerCustomFieldKey(name)
		if key == "" {
			continue
		}

		cfMap, err := jsonquery.Convertor.ObjectToMap(cfNode)
		if err != nil {
			return nil, err
		}

		if v, ok := cfMap["value"]; ok {
			root[key] = v
		}
	}

	return root, nil
}

func phoneburnerCustomFieldValueType(typeName string) common.ValueType {
	switch strings.TrimSpace(strings.ToLower(typeName)) {
	case "text field":
		return common.ValueTypeString
	case "number field":
		return common.ValueTypeFloat
	case "date field":
		return common.ValueTypeDate
	case "check box":
		return common.ValueTypeBoolean
	case "drop down":
		return common.ValueTypeSingleSelect
	default:
		return common.ValueTypeOther
	}
}


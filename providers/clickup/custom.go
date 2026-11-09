package clickup

import (
	"strings"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func recordTransformerWithCustomFields(node *ajson.Node) (map[string]any, error) {
	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	flattenContactCustomFieldsInMap(record)

	return record, nil
}

func flattenContactCustomFieldsInMap(record map[string]any) map[string]any {
	raw, ok := record["custom_fields"]
	if !ok || raw == nil {
		return record
	}

	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		delete(record, "custom_fields")

		return record
	}

	for _, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := obj["name"].(string)
		if name == "" {
			continue
		}

		key := customFieldMetadataKey(name)
		if v, exists := obj["value"]; exists {
			record[key] = v
		}
	}

	delete(record, "custom_fields")

	return record
}

// customFieldMetadataKey is the connector field name for a ClickUp custom field: the provider
// display_name / custom_fields[].name lowercased with spaces preserved
// (e.g. "Next Step Summary" -> "next step summary").
func customFieldMetadataKey(displayName string) string {
	return strings.ToLower(strings.TrimSpace(displayName))
}

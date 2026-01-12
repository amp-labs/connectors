package stripe

import (
	"maps"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// flattenCustomFields converts a Stripe record node into a flat map and promotes keys
// from the nested "metadata" object to the root level. This is used so that
// user-defined custom fields keys can be queried directly as top-level fields in
// ReadResult.Fields, while leaving ReadResult.Raw untouched.
func flattenCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	// custom fields are stored in the metadata object
	customFieldsValue, ok := root["metadata"]
	if !ok {
		// No metadata field, return as is
		return root, nil
	}

	customFieldsMap, ok := customFieldsValue.(map[string]any)
	if !ok || len(customFieldsMap) == 0 {
		// custom fields is not a map or is empty, return as is
		return root, nil
	}

	// flatten metadata keys to root level
	maps.Copy(root, customFieldsMap)

	return root, nil
}

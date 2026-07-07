package stripe

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getCustomFields converts a Stripe record node into a flat map and returns custom fields
// from the nested "metadata" object. This is used so that user-defined
// custom fields keys can be queried directly as top-level fields ReadResult.Fields.
func getCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	// custom fields are stored in the metadata object
	customFieldsValue, ok := root["metadata"]
	if !ok {
		// No metadata field, return nothing.
		return make(map[string]any), nil
	}

	customFieldsMap, ok := customFieldsValue.(map[string]any)
	if !ok || len(customFieldsMap) == 0 {
		// Custom fields is not a map or is empty, return nothing
		return make(map[string]any), nil
	}

	return customFieldsMap, nil
}

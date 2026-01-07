package stripe

import (
	"maps"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// flattenMetadata converts a Stripe record node into a flat map and promotes keys
// from the nested "metadata" object to the root level. This is used so that
// user-defined metadata keys can be queried directly as top-level fields in
// ReadResult.Fields, while leaving ReadResult.Raw untouched.
//
// If a metadata key conflicts with an existing root-level key, the root-level
// key takes precedence and the conflicting metadata entry is ignored. This
// preserves data integrity and ensures Fields can be safely used for write operations.
func flattenMetadata(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	metadataValue, ok := root["metadata"]
	if !ok {
		// No metadata field, return as is
		return root, nil
	}

	metadataMap, ok := metadataValue.(map[string]any)
	if !ok || len(metadataMap) == 0 {
		// metadata is not a map or is empty, return as is
		return root, nil
	}

	// flatten metadata keys to root level
	maps.Copy(root, metadataMap)

	return root, nil
}

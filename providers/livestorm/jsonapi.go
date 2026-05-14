package livestorm

import (
	"maps"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// extractJSONAPIResourceNodes returns JSON:API primary `data` resource nodes only (no flattening).
// List endpoints use `data` as an array (the usual Read path). Livestorm also returns `data` as a
// single resource object for GET /v1/jobs/{id}; that yields one row. If Read is narrowed to
// collections-only later, drop jobs (or similar singleton reads) from read routing instead of
// changing this extractor.
func extractJSONAPIResourceNodes(root *ajson.Node) ([]*ajson.Node, error) {
	items, arrErr := jsonquery.New(root).ArrayOptional("data")
	if arrErr == nil {
		return jsonAPIDataArrayToObjectNodes(items), nil
	}

	obj, objErr := jsonquery.New(root).ObjectOptional("data")
	if objErr == nil && obj != nil {
		if !obj.IsObject() {
			return []*ajson.Node{}, nil
		}

		return []*ajson.Node{obj}, nil
	}

	if arrErr != nil && objErr != nil {
		return nil, arrErr
	}

	return []*ajson.Node{}, nil
}

func jsonAPIDataArrayToObjectNodes(items []*ajson.Node) []*ajson.Node {
	out := make([]*ajson.Node, 0, len(items))

	for _, n := range items {
		if n == nil || !n.IsObject() {
			continue
		}

		out = append(out, n)
	}

	return out
}

// flattenJSONAPIResourceForFields merges JSON:API `id` and `attributes` into one map for
// readhelper.MakeMarshaledDataFuncWithId field selection only. Raw rows use the full node map.
func flattenJSONAPIResourceForFields(n *ajson.Node) (map[string]any, error) {
	id, err := jsonquery.New(n).StringRequired("id")
	if err != nil {
		return nil, err
	}

	attrs, err := jsonquery.New(n).ObjectOptional("attributes")
	if err != nil {
		return nil, err
	}

	merged := map[string]any{"id": id}

	if attrs == nil {
		return merged, nil
	}

	attrMap, err := jsonquery.Convertor.ObjectToMap(attrs)
	if err != nil {
		return nil, err
	}

	maps.Copy(merged, attrMap)

	return merged, nil
}

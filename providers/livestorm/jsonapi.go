package livestorm

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// extractJSONAPIDataRecords returns flattened records from a JSON:API body.
// The "data" property may be an array of resources or a single resource object.
func extractJSONAPIDataRecords(root *ajson.Node) ([]map[string]any, error) {
	items, arrErr := jsonquery.New(root).ArrayOptional("data")
	if arrErr == nil {
		return flattenJSONAPIResourceList(items)
	}

	obj, objErr := jsonquery.New(root).ObjectOptional("data")
	if objErr == nil && obj != nil {
		flat, err := flattenJSONAPIResource(obj)
		if err != nil {
			return nil, err
		}

		return []map[string]any{flat}, nil
	}

	if arrErr != nil && objErr != nil {
		return nil, arrErr
	}

	return []map[string]any{}, nil
}

func flattenJSONAPIResourceList(items []*ajson.Node) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(items))

	for _, n := range items {
		if n == nil {
			continue
		}

		if !n.IsObject() {
			continue
		}

		m, err := flattenJSONAPIResource(n)
		if err != nil {
			return nil, err
		}

		out = append(out, m)
	}

	return out, nil
}

// flattenJSONAPIResource merges resource id and attributes into one map (JSON:API style).
func flattenJSONAPIResource(n *ajson.Node) (map[string]any, error) {
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

	for k, v := range attrMap {
		merged[k] = v
	}

	return merged, nil
}

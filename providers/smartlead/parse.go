package smartlead

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).Array("", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func getNextRecordsURL(_ *ajson.Node) (string, error) {
	// Pagination is not supported for this provider.
	return "", nil
}

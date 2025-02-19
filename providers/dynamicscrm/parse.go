package dynamicscrm

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).ArrayRequired("value")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("@odata.nextLink", "")
}

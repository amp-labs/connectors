package gong

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeGetTotalSize(objectName string) common.ListSizeFunc {
	return func(node *ajson.Node) (int64, error) {
		return jsonquery.New(node).ArraySize(objectName)
	}
}

func makeGetRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// items are stored in array named after the API object
		arr, err := jsonquery.New(node).Array(objectName, false)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

// getNextRecords returns the token or empty string if there are no more records.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "records").StrWithDefault("cursor", "")
}
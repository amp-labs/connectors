package salesforce

import (
	"errors"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getTotalSize returns the total number of records that match the query.
func getTotalSize(node *ajson.Node) (int64, error) {
	size, err := jsonquery.New(node).Integer("totalSize", false)
	if err != nil {
		if !errors.Is(err, jsonquery.ErrKeyNotFound) {
			return 0, err
		}

		// The totalSize key was missing. Try to manually count the number of records
		return jsonquery.New(node).ArraySize("records")
	}

	return *size, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := jsonquery.New(node).Array("records", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
}

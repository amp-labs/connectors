package salesforce

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

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

package outreach

import (
	"encoding/json"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the "next" url for the next page of results,
// If available, else returns an empty string.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	nextPageURL, err := jsonquery.New(node, "links").StringOptional("next")
	if err != nil {
		return "", err
	}

	if nextPageURL == nil {
		return "", nil
	}

	return *nextPageURL, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	var d Data

	b := node.Source()
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}

	records := constructRecords(d)

	return records, nil
}

func constructRecords(d Data) []map[string]any {
	records := make([]map[string]any, len(d.Data))

	for idx, record := range d.Data {
		recordItems := make(map[string]any)
		recordItems[idKey] = record.ID

		// Attributes are flattened into the recordItems map.
		for k, v := range record.Attributes {
			recordItems[k] = v
		}

		recordItems[relationshipsKey] = record.Relationships

		records[idx] = recordItems
	}

	return records
}

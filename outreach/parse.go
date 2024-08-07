package outreach

import (
	"encoding/json"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the "next" url for the next page of results,
// If available, else returns an empty string.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	nextPageURL, err := jsonquery.New(node, "links").Str("next", true)
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

func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("data")
}

// getMarshalledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func getMarshalledData(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

func constructRecords(d Data) []map[string]any {
	records := make([]map[string]any, len(d.Data))

	for i, record := range d.Data {
		recordItems := make(map[string]any)
		recordItems[idKey] = record.ID

		for k, v := range record.Attributes {
			recordItems[k] = v
		}

		records[i] = recordItems
	}

	return records
}

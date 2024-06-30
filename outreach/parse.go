package outreach

import (
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
	records, err := jsonquery.New(node).Array("data", true)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("data")
}

// getMarshaledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func getMarshaledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

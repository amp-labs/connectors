package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageToken", "")
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	result, err := jsonquery.New(node).Array("result", true)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(result)
}

func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("result")
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

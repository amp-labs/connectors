package g2

import (
	"encoding/json"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

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

// getMarshalledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func getMarshalledData(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	var recordId string

	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		id, exists := record["id"]
		if exists {
			idStr, ok := id.(string)
			if ok {
				recordId = idStr
			}
		}

		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
			Id:     recordId,
		}
	}

	return data, nil
}

func constructRecords(d Data) []map[string]any {
	records := make([]map[string]any, len(d.Data))

	for i, record := range d.Data {
		recordItems := make(map[string]any)
		recordItems["id"] = record.ID

		maps.Copy(recordItems, record.Attributes)

		records[i] = recordItems
	}

	return records
}

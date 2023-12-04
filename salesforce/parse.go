package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("records")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, ErrNotArray
	}

	arr := records.MustArray()

	out := make([]map[string]interface{}, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, ErrNotObject
		}

		out = append(out, m)
	}

	return out, nil
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("nextRecordsUrl") {
		next, err := node.GetKey("nextRecordsUrl")
		if err != nil {
			return "", err
		}

		if !next.IsString() {
			return "", ErrNotString
		}

		nextPage = next.MustString()
	}

	return nextPage, nil
}

// getTotalSize returns the total number of records that match the query.
func getTotalSize(node *ajson.Node) (int64, error) {
	node, err := node.GetKey("totalSize")
	if err != nil {
		return 0, err
	}

	if !node.IsNumeric() {
		return 0, ErrNotNumeric
	}

	return int64(node.MustNumeric()), nil
}

func getStructuredData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

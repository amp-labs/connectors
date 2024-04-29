package gong

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the token or empty string if there are no more records.
func getNextRecordsURL(node *ajson.Node) (string, error) {

	recordsNode, err := node.GetKey("records")
	if err != nil {
		return "", err
	}

	cursorNode, err := recordsNode.GetKey("cursor")
	if err != nil {
		return "", err
	}

	nextPage := cursorNode.String()
	return nextPage, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node, objectName string) ([]map[string]interface{}, error) {

	records, err := node.GetKey(objectName)
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

// getTotalSize returns the total number of records that match the query.
func getTotalSize(node *ajson.Node) (int64, error) {

	recordsNode, err := node.GetKey("records")
	if err != nil {
		return 0, err
	}

	totalRecordsNode, err := recordsNode.GetKey("totalRecords")
	if err != nil {
		return 0, err
	}

	totalRecords, err := totalRecordsNode.GetNumeric()
	if err != nil {
		return 0, err
	}

	return int64(totalRecords), nil

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

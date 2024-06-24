package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// getNextRecordsAfter returns the "after" value for the next page of results.
func getNextRecordsAfter(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		after, err := next.GetKey("after")
		if err != nil {
			return "", err
		}

		if !after.IsString() {
			return "", ErrNotString
		}

		nextPage = after.MustString()
	}

	return nextPage, nil
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		link, err := next.GetKey("link")
		if err != nil {
			return "", err
		}

		if !link.IsString() {
			return "", ErrNotString
		}

		nextPage = link.MustString()
	}

	return nextPage, nil
}

// parsePagingNext is a helper to return the paging.next node.
func parsePagingNext(node *ajson.Node) (*ajson.Node, error) {
	paging, err := node.GetKey("paging")
	if err != nil {
		return nil, err
	}

	if !paging.IsObject() {
		return nil, ErrNotObject
	}

	next, err := paging.GetKey("next")
	if err != nil {
		return nil, err
	}

	if !next.IsObject() {
		return nil, ErrNotObject
	}

	return next, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("results")
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
	node, err := node.GetKey("results")
	if err != nil {
		return 0, err
	}

	if !node.IsArray() {
		return 0, ErrNotArray
	}

	return int64(node.Size()), nil
}

// getMarshaledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func getMarshaledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		recordProperties, ok := record["properties"].(map[string]interface{})
		if !ok {
			return nil, ErrNotObject
		}

		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, recordProperties),
			Raw:    record,
		}
	}

	return data, nil
}

// GetLastResultId returns the last row's id from a result.
func GetLastResultId(result *common.ReadResult) string {
	numRecords := len(result.Data)
	if numRecords == 0 {
		return ""
	}

	// Get the last row and get the hs_object_id field's value
	lastRow := result.Data[numRecords-1]

	lastRowId, ok := lastRow.Fields[string(ObjectFieldHsObjectId)].(string)
	if !ok {
		return ""
	}

	return lastRowId
}

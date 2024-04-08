package outreach

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the "next" url for the next page of results,
// If available, else returns an empty string
func getNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("links") {
		links, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		if links.HasKey("next") {
			next, err := links.GetKey("next")
			if err != nil {
				return "", err
			}

			if !next.IsString() {
				return "", ErrNotString
			}

			nextPage = next.MustString()
		}
	}
	return nextPage, nil
}

// parsePagingNext is a helper to return the links node.
func parsePagingNext(node *ajson.Node) (*ajson.Node, error) {
	links, err := node.GetKey("links")
	if err != nil {
		return nil, err
	}

	if !links.IsObject() {
		return nil, ErrNotObject
	}

	return links, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("data")
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
	node, err := node.GetKey("data")
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
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

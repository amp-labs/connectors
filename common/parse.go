package common

import (
	"github.com/spyzhov/ajson"
)

// ParseResult parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc, recordsFunc, nextPageFunc, and structureFunc are used to extract the relevant data from the response.
// The sizeFunc returns the total number of records in the response.
// The recordsFunc returns the records in the response.
// The nextPageFunc returns the URL for the next page of results.
// The structureFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResult(
	data *ajson.Node,
	sizeFunc func(*ajson.Node) (int64, error),
	recordsFunc func(*ajson.Node) ([]map[string]any, error),
	nextPageFunc func(*ajson.Node) (string, error),
	structureFunc func([]map[string]any, []string) ([]ReadResultRow, error),
	fields []string,
) (*ReadResult, error) {
	totalSize, err := sizeFunc(data)
	if err != nil {
		return nil, err
	}

	records, err := recordsFunc(data)
	if err != nil {
		return nil, err
	}

	nextPage, err := nextPageFunc(data)
	if err != nil {
		return nil, err
	}

	done := nextPage == ""

	structuredData, err := structureFunc(records, fields)
	if err != nil {
		return nil, err
	}

	return &ReadResult{
		Rows:     totalSize,
		Data:     structuredData,
		NextPage: NextPageToken(nextPage),
		Done:     done,
	}, nil
}

// ExtractFieldsFromRaw returns a map of fields from a record.
func ExtractFieldsFromRaw(fields []string, record map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(fields))

	for _, field := range fields {
		if value, ok := record[field]; ok {
			out[field] = value
		}
	}

	return out
}

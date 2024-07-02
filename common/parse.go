package common

import (
	"strings"

	"github.com/spyzhov/ajson"
)

type (
	ListSizeFunc func(*ajson.Node) (int64, error)
	NextPageFunc func(*ajson.Node) (string, error)
	RecordsFunc  func(*ajson.Node) ([]map[string]any, error)
)

// ParseResult parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc, recordsFunc, nextPageFunc, and marshalFunc are used to extract the relevant data from the response.
// The sizeFunc returns the total number of records in the response.
// The recordsFunc returns the records in the response.
// The nextPageFunc returns the URL for the next page of results.
// The marshalFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResult(
	resp *JSONHTTPResponse,
	sizeFunc func(*ajson.Node) (int64, error),
	recordsFunc func(*ajson.Node) ([]map[string]any, error),
	nextPageFunc func(*ajson.Node) (string, error),
	marshalFunc func([]map[string]any, []string) ([]ReadResultRow, error),
	fields []string,
) (*ReadResult, error) {
	if resp == nil {
		return nil, ErrEmptyJSONHTTPResponse
	}

	totalSize, err := sizeFunc(resp.Body)
	if err != nil {
		return nil, err
	}

	records, err := recordsFunc(resp.Body)
	if err != nil {
		return nil, err
	}

	nextPage, err := nextPageFunc(resp.Body)
	if err != nil {
		return nil, err
	}

	done := nextPage == ""

	marshaledData, err := marshalFunc(records, fields)
	if err != nil {
		return nil, err
	}

	return &ReadResult{
		Rows:     totalSize,
		Data:     marshaledData,
		NextPage: NextPageToken(nextPage),
		Done:     done,
	}, nil
}

// ExtractLowercaseFieldsFromRaw returns a map of fields from a record.
// The fields are all returned in lowercase.
func ExtractLowercaseFieldsFromRaw(fields []string, record map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(fields))

	// Modify all record keys to lowercase
	lowercaseRecord := make(map[string]interface{}, len(record))
	for key, value := range record {
		lowercaseRecord[strings.ToLower(key)] = value
	}

	for _, field := range fields {
		// Lowercase the field name to make lookup case-insensitive.
		lowercaseField := strings.ToLower(field)

		if value, ok := lowercaseRecord[lowercaseField]; ok {
			out[lowercaseField] = value
		}
	}

	return out
}

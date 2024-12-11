package common

import (
	"strings"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

type (
	NextPageFunc func(*ajson.Node) (string, error)
	RecordsFunc  func(*ajson.Node) ([]map[string]any, error)
)

// ParseResult parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc returns the total number of records in the response.
// The recordsFunc returns the records in the response.
// The nextPageFunc returns the URL for the next page of results.
// The marshalFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResult(
	resp *JSONHTTPResponse,
	recordsFunc func(*ajson.Node) ([]map[string]any, error),
	nextPageFunc func(*ajson.Node) (string, error),
	marshalFunc func([]map[string]any, []string) ([]ReadResultRow, error),
	fields datautils.Set[string],
) (*ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, ErrEmptyJSONHTTPResponse
	}

	records, err := recordsFunc(body)
	if err != nil {
		return nil, err
	}

	nextPage, err := nextPageFunc(body)
	if err != nil {
		return nil, err
	}

	marshaledData, err := marshalFunc(records, fields.List())
	if err != nil {
		return nil, err
	}

	// Next page doesn't exist if:
	// * either there is no next page token,
	// * or current page was empty.
	// This will guarantee that Read is finite.
	done := nextPage == "" || len(marshaledData) == 0
	if done {
		// It is possible that the provider doesn't reset the next page token when there are no more records.
		// In this case, we should set the next page token to an empty string to indicate that we are done.
		nextPage = ""
	}

	return &ReadResult{
		Rows:     int64(len(marshaledData)),
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

func GetMarshaledData(records []map[string]any, fields []string) ([]ReadResultRow, error) {
	data := make([]ReadResultRow, len(records))

	for i, record := range records {
		data[i] = ReadResultRow{
			Fields: ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

func GetRecordsUnderJSONPath(jsonPath string) RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := jsonquery.New(node).Array(jsonPath, false)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func GetOptionalRecordsUnderJSONPath(jsonPath string) RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := jsonquery.New(node).Array(jsonPath, true)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

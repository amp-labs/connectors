package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

type (
	// NextPageFunc extracts the next page token/URL from the response body.
	NextPageFunc func(*ajson.Node) (string, error)

	// RecordsFunc extracts a list of records as map[string]any from the response body.
	RecordsFunc func(*ajson.Node) ([]map[string]any, error)
	// NodeRecordsFunc extracts a list of records as ajson.Node from the response body.
	NodeRecordsFunc func(*ajson.Node) ([]*ajson.Node, error)

	// RawReadRecordFunc processes a raw record, applying additional transformations if needed.
	// There is one common example of flattening nested fields provided by RecordFlattener.
	RawReadRecordFunc func(node *ajson.Node) (map[string]any, error)
	// MarshalFunc converts a list of map[string]any records into ReadResultRow format.
	MarshalFunc func(records []map[string]any, fields []string) ([]ReadResultRow, error)
	// NodeMarshalFunc converts a list of ajson.Node records into ReadResultRow format.
	NodeMarshalFunc func(records []*ajson.Node, fields []string) ([]ReadResultRow, error)
)

// RawReadRecord defines the types of records that ParseResult can process.
// It determines which callback function is used for parsing.
type RawReadRecord interface {
	map[string]any | *ajson.Node
}

// ParseResult parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc returns the total number of records in the response.
// The recordsFunc returns the records in the response.
// The nextPageFunc returns the URL for the next page of results.
// The marshalFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResult[R RawReadRecord](
	resp *JSONHTTPResponse,
	recordsFunc func(*ajson.Node) ([]R, error),
	nextPageFunc func(*ajson.Node) (string, error),
	marshalFunc func([]R, []string) ([]ReadResultRow, error),
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

	if len(marshaledData) == 0 {
		// Either a JSON array is empty or it was nil.
		// For consistency return empty array for missing records.
		marshaledData = make([]ReadResultRow, 0)
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
func ExtractLowercaseFieldsFromRaw(fields []string, record map[string]any) map[string]any {
	out := make(map[string]any, len(fields))

	// Modify all record keys to lowercase
	lowercaseRecord := make(map[string]any, len(record))
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

var (
	errMissingId        = errors.New("missing id field in raw record")
	errUnexpectedIdType = errors.New("unexpected id type")
)

// GetMarshalledDataWithId is very similar to GetMarshaledData, but it also extracts the "id" field from the raw record.
func GetMarshalledDataWithId(records []map[string]any, fields []string) ([]ReadResultRow, error) {
	data := make([]ReadResultRow, len(records))

	fields = append(fields, "id")

	//nolint:varnamelen
	for i, record := range records {
		data[i] = ReadResultRow{
			Fields: ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}

		idAny := data[i].Fields["id"]
		if idAny == nil {
			return nil, errMissingId
		}

		id, ok := idAny.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %T", errUnexpectedIdType, idAny)
		}

		data[i].Id = id
	}

	return data, nil
}

// MakeMarshaledDataFunc produces ReadResultRow where raw record differs from the fields.
// This usually includes a set of actions to preprocess, usually to flatten the raw record and then extract
// fields requested by the user.
func MakeMarshaledDataFunc(nodeRecordFunc RawReadRecordFunc) NodeMarshalFunc {
	return func(records []*ajson.Node, fields []string) ([]ReadResultRow, error) {
		data := make([]ReadResultRow, len(records))

		for index, nodeRecord := range records {
			raw, err := jsonquery.Convertor.ObjectToMap(nodeRecord)
			if err != nil {
				return nil, err
			}

			record, err := nodeRecordFunc(nodeRecord)
			if err != nil {
				return nil, err
			}

			data[index] = ReadResultRow{
				Fields: ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    raw,
			}
		}

		return data, nil
	}
}

func GetRecordsUnderJSONPath(jsonPath string, nestedPath ...string) func(*ajson.Node) ([]map[string]any, error) {
	return getRecords(false, jsonPath, nestedPath...)
}

func GetOptionalRecordsUnderJSONPath(jsonPath string, nestedPath ...string) RecordsFunc {
	return getRecords(true, jsonPath, nestedPath...)
}

func getRecords(optional bool, jsonPath string, nestedPath ...string) RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		var (
			arr []*ajson.Node
			err error
		)

		if optional {
			arr, err = jsonquery.New(node, nestedPath...).ArrayOptional(jsonPath)
		} else {
			arr, err = jsonquery.New(node, nestedPath...).ArrayRequired(jsonPath)
		}

		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

// RecordFlattener returns a procedure which copies fields of a nested object to the top level.
//
// Ex: Every object has special field "attributes" which holds all the object specific fields.
// Therefore, nested "attributes" will be removed and fields will be moved to the top level of the object.
//
// Example accounts(shortened response):
//
//	 "data": [
//	    {
//	        "type": "",
//	        "id": "",
//	        "attributes": {
//	            "test_account": false,
//	            "contact_information": {},
//	            "locale": ""
//	        },
//	        "links": {}
//	    }
//	],
//
// The resulting fields for the above will be: [ type, id, test_account, contact_information, locale, links ].
func RecordFlattener(nestedKey string) RawReadRecordFunc {
	return func(node *ajson.Node) (map[string]any, error) {
		attributes, err := jsonquery.New(node).ObjectOptional(nestedKey)
		if err != nil {
			return nil, err
		}

		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		nested, err := jsonquery.Convertor.ObjectToMap(attributes)
		if err != nil {
			return nil, err
		}

		// Nested object will be removed.
		delete(root, nestedKey)

		// Fields from attributes are moved to the top level.
		for key, value := range nested {
			root[key] = value
		}

		// Root level has adopted fields from nested object.
		return root, nil
	}
}

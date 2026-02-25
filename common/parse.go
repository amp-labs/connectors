// nolint:revive,godoclint
package common

import (
	"encoding/json"
	"maps"
	"strconv"
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

	// RecordsFilterFunc filters records based on ReadParams (e.g., Since/Until ranges).
	// It returns:
	//   - A filtered slice of records
	//   - A next page token (which may be empty if no more records are available using deduction)
	//   - An error, if filtering fails
	RecordsFilterFunc func(ReadParams, *ajson.Node, []*ajson.Node) ([]*ajson.Node, string, error)

	// RecordTransformer is a function that processes a JSON node and transforms it
	// into a map representation, potentially applying structural modifications.
	//
	// Common use cases include:
	// - Flattening nested objects (see FlattenNestedFields)
	// - Filtering out unwanted fields
	// - Converting field formats
	// - Renaming fields
	// - Adding computed fields
	//
	// Example usage:
	// - FlattenNestedFields("attributes")
	// - Replacing GUIDs with human-readable fields
	// - Enhancing fields with custom data from the API response.
	// - Adding a relationships property to the root level from a nested object.
	RecordTransformer func(node *ajson.Node) (map[string]any, error)
	// MarshalFunc converts a list of map[string]any records into ReadResultRow format.
	MarshalFunc func(records []map[string]any, fields []string) ([]ReadResultRow, error)
	// MarshalFromNodeFunc converts a list of ajson.Node records into ReadResultRow format.
	MarshalFromNodeFunc func(records []*ajson.Node, fields []string) ([]ReadResultRow, error)
)

// ProviderReadResponseType defines the types of records that ParseResult can process.
// It determines which callback function is used for parsing.
type ProviderReadResponseType interface {
	map[string]any | *ajson.Node
}

// ParseResult parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc returns the total number of records in the response.
// The extractRecords returns the records in the response.
// The extractNextPage returns the URL for the next page of results.
// The marshalFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResult[R ProviderReadResponseType](
	resp *JSONHTTPResponse,
	extractRecords func(*ajson.Node) ([]R, error),
	extractNextPage func(*ajson.Node) (string, error),
	marshalFunc func([]R, []string) ([]ReadResultRow, error),
	fields datautils.Set[string],
) (*ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, ErrEmptyJSONHTTPResponse
	}

	records, err := extractRecords(body)
	if err != nil {
		return nil, err
	}

	nextPage, err := extractNextPage(body)
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

// ParseResultFiltered parses the response from a provider into a ReadResult. A 2xx return type is assumed.
// The sizeFunc returns the total number of records in the response.
// The extractRecords returns the records in the response.
// The filterRecords acts as a sieve based on ReadParams since and until properties.
// The marshalFunc is used to structure the data into an array of ReadResultRows.
// The fields are used to populate ReadResultRow.Fields.
func ParseResultFiltered(
	params ReadParams,
	resp *JSONHTTPResponse,
	extractRecords func(*ajson.Node) ([]*ajson.Node, error),
	filterRecords func(ReadParams, *ajson.Node, []*ajson.Node) ([]*ajson.Node, string, error),
	marshalFunc func([]*ajson.Node, []string) ([]ReadResultRow, error),
	fields datautils.Set[string],
) (*ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, ErrEmptyJSONHTTPResponse
	}

	unfilteredRecords, err := extractRecords(body)
	if err != nil {
		return nil, err
	}

	records, nextPage, err := filterRecords(params, body, unfilteredRecords)
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

// GetMarshaledData extracts the "id" field from the raw record and returns a list of ReadResultRow.
func GetMarshaledData(records []map[string]any, fields []string) ([]ReadResultRow, error) {
	data := make([]ReadResultRow, len(records))

	fields = append(fields, "id")

	//nolint:varnamelen
	for i, record := range records {
		data[i] = ReadResultRow{
			Fields: ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}

		var id string

		switch v := data[i].Fields["id"].(type) {
		case string:
			id = v
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
		case json.Number:
			id = v.String()
		}

		data[i].Id = id
	}

	return data, nil
}

// MakeMarshaledDataFunc constructs a MarshalFromNodeFunc that converts records into ReadResultRow slices.
// It applies an optional RecordTransformer to each record; if nil, it defaults to ajson-to-map conversion.
// Typically used to flatten, normalize records or enhance them with custom fields.
func MakeMarshaledDataFunc(nodeRecordFunc RecordTransformer) MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]ReadResultRow, error) {
		if nodeRecordFunc == nil {
			// Default method converts ajson.Node to map[string]any.
			// If conversion is not enough and data should be altered
			// non-nil RecordTransformer should've been given.
			nodeRecordFunc = func(node *ajson.Node) (map[string]any, error) {
				return jsonquery.Convertor.ObjectToMap(node)
			}
		}

		fields = append(fields, "id")
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

			extractedFields := ExtractLowercaseFieldsFromRaw(fields, record)

			var recordID string

			switch v := extractedFields["id"].(type) {
			case string:
				recordID = v
			case float64:
				recordID = strconv.FormatFloat(v, 'f', -1, 64)
			case json.Number:
				recordID = v.String()
			}

			data[index] = ReadResultRow{
				Fields: extractedFields,
				Raw:    raw,
				Id:     recordID,
			}
		}

		return data, nil
	}
}

// MakeRecordsFunc returns a function that extracts an array of record nodes from a JSON document.
// The jsonPath defines where the records array resides.
// The optional nestedPath specifies a traversal prefix to locate nested arrays.
func MakeRecordsFunc(jsonPath string, nestedPath ...string) NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node, nestedPath...).ArrayRequired(jsonPath)
	}
}

func ExtractRecordsFromPath(jsonPath string, nestedPath ...string) RecordsFunc {
	return extractRecords(false, jsonPath, nestedPath...)
}

func ExtractOptionalRecordsFromPath(jsonPath string, nestedPath ...string) RecordsFunc {
	return extractRecords(true, jsonPath, nestedPath...)
}

func extractRecords(optional bool, jsonPath string, nestedPath ...string) RecordsFunc {
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

// FlattenNestedFields returns a procedure which copies fields of a nested object to the top level.
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
func FlattenNestedFields(nestedKey string) RecordTransformer {
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
		maps.Copy(root, nested)

		// Root level has adopted fields from nested object.
		return root, nil
	}
}

// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"maps"
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

// GetMarshaledData converts records into ReadResultRow slices without populating the Id field.
//
// Deprecated: Use MakeGetMarshaledDataWithId instead, which supports Id extraction
// with configurable field mappings for both flat and nested ID structures.
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

// IdFieldQuery specifies how to locate an ID field within a record.
// For flat structures where the ID is at the root level, only Field is needed.
// For nested structures, Zoom provides the path of keys to traverse before accessing Field.
//
// Examples:
//
//	Flat ID at root:       IdFieldQuery{Field: "id"}           -> record["id"]
//	Nested ID:             IdFieldQuery{Zoom: []string{"id"}, Field: "record_id"}
//	                                                           -> record["id"]["record_id"]
//	Deeply nested:         IdFieldQuery{Zoom: []string{"meta", "info"}, Field: "uid"}
//	                                                           -> record["meta"]["info"]["uid"]
type IdFieldQuery struct {
	// Zoom is the path of keys to traverse to reach the nested object containing the ID.
	// If nil or empty, the ID field is expected at the root level of the record.
	Zoom []string
	// Field is the name of the field containing the ID value.
	Field string
}

// NewIdField creates an IdFieldQuery for a flat ID field at the root level.
// This is the common case where the ID is directly accessible as record[field].
func NewIdField(field string) IdFieldQuery {
	return IdFieldQuery{Field: field}
}

// NewNestedIdField creates an IdFieldQuery for an ID field nested within the record.
// The zoom path specifies the keys to traverse, and field is the final key containing the ID.
//
// Example: For a record like {"id": {"record_id": "123"}}, use:
//
//	NewNestedIdField([]string{"id"}, "record_id")
func NewNestedIdField(zoom []string, field string) IdFieldQuery {
	return IdFieldQuery{Zoom: zoom, Field: field}
}

// MakeGetMarshaledDataWithId constructs a MarshalFunc that converts records into ReadResultRow slices
// with the Id field populated. It uses the provided idFieldMapping to determine which field
// contains the record ID for each object type.
//
// The idFieldMapping should be a DefaultMap that returns the IdFieldQuery for a given object name.
// Most APIs use "id" as the default, so a typical mapping would be:
//
//	idFieldMapping := datautils.NewDefaultMap(datautils.Map[string, IdFieldQuery]{
//	    "specialObject": NewNestedIdField([]string{"id"}, "record_id"),  // nested ID
//	}, func(_ string) IdFieldQuery { return NewIdField("id") })  // default: flat "id"
//
// This function gracefully handles missing ID fields by leaving ReadResultRow.Id empty,
// allowing backwards compatibility with existing connectors.
func MakeGetMarshaledDataWithId(
	objectName string,
	idFieldMapping datautils.DefaultMap[string, IdFieldQuery],
) MarshalFunc {
	return func(records []map[string]any, fields []string) ([]ReadResultRow, error) {
		idQuery := idFieldMapping.Get(objectName)
		data := make([]ReadResultRow, len(records))

		for i, record := range records {
			data[i] = ReadResultRow{
				Fields: ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    record,
				Id:     extractIdFromRecord(record, idQuery),
			}
		}

		return data, nil
	}
}

// MakeMarshaledDataFuncWithId constructs a MarshalFromNodeFunc that converts records into ReadResultRow slices
// with the Id field populated. It combines the functionality of MakeMarshaledDataFunc with ID extraction.
//
// The nodeRecordFunc parameter allows custom record transformation (e.g., flattening nested fields).
// If nil, it defaults to standard ajson-to-map conversion.
//
// The idFieldMapping should be a DefaultMap that returns the IdFieldQuery for a given object name.
// Most APIs use "id" as the default, so a typical mapping would be:
//
//	idFieldMapping := datautils.NewDefaultMap(datautils.Map[string, IdFieldQuery]{
//	    "specialObject": NewNestedIdField([]string{"id"}, "record_id"),  // nested ID
//	}, func(_ string) IdFieldQuery { return NewIdField("id") })  // default: flat "id"
//
// This function gracefully handles missing ID fields by leaving ReadResultRow.Id empty,
// allowing backwards compatibility with existing connectors.
func MakeMarshaledDataFuncWithId(
	nodeRecordFunc RecordTransformer,
	objectName string,
	idFieldMapping datautils.DefaultMap[string, IdFieldQuery],
) MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]ReadResultRow, error) {
		idQuery := idFieldMapping.Get(objectName)

		if nodeRecordFunc == nil {
			// Default method converts ajson.Node to map[string]any.
			// If conversion is not enough and data should be altered
			// non-nil RecordTransformer should've been given.
			nodeRecordFunc = func(node *ajson.Node) (map[string]any, error) {
				return jsonquery.Convertor.ObjectToMap(node)
			}
		}

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
				Id:     extractIdFromRecord(raw, idQuery),
			}
		}

		return data, nil
	}
}

// extractIdFromRecord extracts the ID value from a record map using the provided IdFieldQuery.
// It traverses the zoom path (if any) to reach nested objects, then extracts the ID field.
// It handles both string and numeric (float64) ID types, converting numbers to strings.
// Returns empty string if the path is invalid, the ID field is missing, or has an unsupported type.
func extractIdFromRecord(record map[string]any, query IdFieldQuery) string {
	// Navigate through the zoom path to reach the nested object containing the ID.
	current := record

	for _, key := range query.Zoom {
		nested, ok := current[key]
		if !ok {
			return ""
		}

		nestedMap, ok := nested.(map[string]any)
		if !ok {
			return ""
		}

		current = nestedMap
	}

	// Extract the ID field from the current (possibly nested) object.
	idValue, exists := current[query.Field]
	if !exists {
		return ""
	}

	switch id := idValue.(type) {
	case string:
		return id
	case float64:
		// JSON numbers are parsed as float64, convert to string without decimal places.
		return fmt.Sprintf("%.0f", id)
	default:
		return ""
	}
}

// MakeMarshaledDataFunc constructs a MarshalFromNodeFunc that converts records into ReadResultRow slices.
// It applies an optional RecordTransformer to each record; if nil, it defaults to ajson-to-map conversion.
// Typically used to flatten, normalize records or enhance them with custom fields.
//
// Deprecated: Use MakeMarshaledDataFuncWithId instead, which supports Id extraction
// with configurable field mappings for both flat and nested ID structures.
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

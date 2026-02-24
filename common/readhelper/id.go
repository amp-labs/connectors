package readhelper

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

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
) common.MarshalFunc {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		idQuery := idFieldMapping.Get(objectName)
		data := make([]common.ReadResultRow, len(records))

		for i, record := range records {
			data[i] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
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
	nodeRecordFunc common.RecordTransformer,
	objectName string,
	idFieldMapping datautils.DefaultMap[string, IdFieldQuery],
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		idQuery := idFieldMapping.Get(objectName)

		if nodeRecordFunc == nil {
			// Default method converts ajson.Node to map[string]any.
			// If conversion is not enough and data should be altered
			// non-nil RecordTransformer should've been given.
			nodeRecordFunc = func(node *ajson.Node) (map[string]any, error) {
				return jsonquery.Convertor.ObjectToMap(node)
			}
		}

		data := make([]common.ReadResultRow, len(records))

		for index, nodeRecord := range records {
			raw, err := jsonquery.Convertor.ObjectToMap(nodeRecord)
			if err != nil {
				return nil, err
			}

			record, err := nodeRecordFunc(nodeRecord)
			if err != nil {
				return nil, err
			}

			data[index] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
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

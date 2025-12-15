package attio

import (
	"maps"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(reqLink *urlbuilder.URL, obj string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		previousStart := 0

		if (reqLink.HasQueryParam("limit") || reqLink.HasQueryParam("offset")) && len(value) != 0 {
			offsetQP, ok := reqLink.GetFirstQueryParam("offset")
			if ok {
				// Try to use previous "offset" parameter to determine the next offset.
				offsetNum, err := strconv.Atoi(offsetQP)
				if err == nil {
					previousStart = offsetNum
				}
			}

			nextStart, pageSize := handlePagination(previousStart, obj)

			reqLink.WithQueryParam("limit", strconv.Itoa(pageSize))
			reqLink.WithQueryParam("offset", strconv.Itoa(nextStart))

			return reqLink.String(), nil
		}

		return "", nil
	}
}

// To determine the next page records for the standard/custom objects.
func makeNextRecordStandardObj(offset int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		if len(value) == 0 {
			return "", nil
		}

		nextStart := offset + DefaultPageSize

		return strconv.Itoa(nextStart), nil
	}
}

func handlePagination(previousStart int, obj string) (int, int) {
	var nextStart, pageSize int

	if obj == objectNameNotes {
		nextStart = previousStart + DefaultPageSizeForNotesObj
		pageSize = DefaultPageSizeForNotesObj
	} else {
		nextStart = previousStart + DefaultPageSize
		pageSize = DefaultPageSize
	}

	return nextStart, pageSize
}

// standard/custom object has a special field named "values" which holds all the important fields.
// Therefore, nested "values" will be removed and fields inside the "values" field will be moved
// to the top level of the object.
//
// Example companies(shortened response):
//
//		"data": [
//		{
//			"id": {
//			  "workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
//			  "object_id": "ffbca575-69c4-4080-bf98-91d79aeea4b1",
//			  "record_id": "d1b0593a-fb43-4d41-82ab-57fc3db73b3a"
//			},
//			"created_at": "2025-03-25T06:44:30.177000000Z",
//			"values": {
//			  "record_id": [
//				{
//				  "active_from": "2025-03-25T06:44:30.177000000Z",
//				  "active_until": null,
//				  "created_by_actor": {
//					"type": "workspace-member",
//					"id": "073f4c74-b60d-4de9-992a-0f799b5e442e"
//				  },
//				  "value": "d1b0593a-fb43-4d41-82ab-57fc3db73b3a",
//				  "attribute_type": "text"
//				}
//			  ],
//	       .... (more response data will be there)
//
// The resulting fields for the above will be: id, created_at, record_id.

type MarshalledData func([]map[string]any, []string) ([]common.ReadResultRow, error)

func DataMarshall(resp *common.JSONHTTPResponse) MarshalledData {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		node, ok := resp.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		arr, err := jsonquery.New(node).ArrayOptional("data")
		if err != nil {
			return nil, err
		}

		flattenrecords, err := flattenRecords(arr)
		if err != nil {
			return nil, err
		}

		return getRecords(flattenrecords, records, fields)
	}
}

func getRecords(
	flattenRecords map[string]any, records []map[string]any, fields []string,
) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records { // nolint:varnamelen
		id, ok := record["id"].(map[string]any) // nolint:varnamelen
		if !ok {
			return nil, common.ErrEmptyRecordIdResponse
		}

		recordID, ok := id["record_id"].(string)
		if !ok {
			return nil, common.ErrEmptyRecordIdResponse
		}

		fieldRecord, ok := flattenRecords[recordID].(map[string]any)
		if !ok {
			return nil, common.ErrEmptyRecordIdResponse
		}

		data[i].Raw = record
		data[i].Fields = common.ExtractLowercaseFieldsFromRaw(fields, fieldRecord)
	}

	return data, nil
}

func flattenRecord(arr *ajson.Node) (map[string]any, error) {
	const keyValuesObject = "values"

	values, err := jsonquery.New(arr).ObjectOptional(keyValuesObject)
	if err != nil {
		return nil, err
	}

	original, err := jsonquery.Convertor.ObjectToMap(arr)
	if err != nil {
		return nil, err
	}

	nested, err := jsonquery.Convertor.ObjectToMap(values)
	if err != nil {
		return nil, err
	}

	// values object must be removed.
	delete(original, keyValuesObject)

	// Fields from values are moved to the top level.
	maps.Copy(original, nested)

	return original, nil
}

func flattenRecords(arr []*ajson.Node) (map[string]any, error) {
	flattenMap := make(map[string]any, 0)

	for _, element := range arr {
		flattenedRecord, err := flattenRecord(element)
		if err != nil {
			return nil, err
		}

		recordId, err := jsonquery.New(element, "id").StringRequired("record_id")
		if err != nil {
			return nil, err
		}

		flattenMap[recordId] = flattenedRecord
	}

	return flattenMap, nil
}

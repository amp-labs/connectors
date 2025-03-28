package attio

import (
	"encoding/json"
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
func makeNextRecordStandardObj(body map[string]any) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		previousStart := 0

		if len(value) != 0 {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return "", err
			}

			// Parse the JSON into an *ajson.Node
			node, err := ajson.Unmarshal(jsonData)
			if err != nil {
				return "", err
			}

			// To determine the offset value.
			offset, err := jsonquery.New(node).IntegerWithDefault("offset", 0)
			if err != nil {
				return "", err
			}

			previousStart = int(offset)

			nextStart := previousStart + DefaultPageSize

			return strconv.Itoa(nextStart), nil
		}

		return "", nil
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

type MarshaledData func([]map[string]any, []string) ([]common.ReadResultRow, error)

func MyGetMarshaledData(resp *common.JSONHTTPResponse) MarshaledData {
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

		return Marshaled(flattenrecords, records, fields)
	}
}

func Marshaled(flattenRecords map[string]map[string]any, records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i := 0; i < len(records); i++ {
		FeildRecord := flattenRecords[records[i]["id"].(map[string]any)["record_id"].(string)]
		data[i].Raw = records[i]
		data[i].Fields = common.ExtractLowercaseFieldsFromRaw(fields, FeildRecord)
	}

	return data, nil
}

func flattenRecords(arr []*ajson.Node) (map[string]map[string]any, error) {
	result := make([]map[string]any, len(arr))

	flattenMap := make(map[string]map[string]any, 0)

	for index, element := range arr {
		const keyValuesObject = "values"

		values, err := jsonquery.New(element).ObjectOptional(keyValuesObject)
		if err != nil {
			return nil, err
		}

		original, err := jsonquery.Convertor.ObjectToMap(element)
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
		for key, value := range nested {
			original[key] = value
		}

		result[index] = original

		flattenMap[original["id"].(map[string]any)["record_id"].(string)] = original
	}

	return flattenMap, nil
}

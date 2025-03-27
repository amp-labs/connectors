package attio

import (
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

			var nextStart int

			reqLink, nextStart = setLimit(previousStart, obj, reqLink)

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
			// To determine the offset value.
			if offset, ok := body["offset"].(int); ok {
				previousStart = offset
			}

			nextStart := previousStart + DefaultPageSize

			return strconv.Itoa(nextStart), nil
		}

		return "", nil
	}
}

func setLimit(previousStart int, obj string, reqLink *urlbuilder.URL) (*urlbuilder.URL, int) {
	var nextStart int

	if obj == objectNameNotes {
		nextStart = previousStart + DefaultPageSizeForNotesObj
		reqLink.WithQueryParam("limit", strconv.Itoa(DefaultPageSizeForNotesObj))
	} else {
		nextStart = previousStart + DefaultPageSize
		reqLink.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
	}

	return reqLink, nextStart
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
func getStandardOrCustomObjRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).ArrayOptional("data")
	if err != nil {
		return nil, err
	}

	return flattenRecords(arr)
}

func flattenRecords(arr []*ajson.Node) ([]map[string]any, error) {
	result := make([]map[string]any, len(arr))

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
	}

	return result, nil
}

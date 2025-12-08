package chorus

import (
	"maps"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const PageSize = 100

// IncrementalObjectQueryParam
// Below objects supports incremental reading but having different query pram values.
// Ref: https://api-docs.chorus.ai/#4dc74394-9852-4b6b-9555-cb8ce951557b for scorecards object.
// Ref: https://api-docs.chorus.ai/#3d962146-73fc-42db-afda-a943971ab1c4 for emails object.
var IncrementalObjectQueryParam = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"emails":     "filter[email.sent]",
	"scorecards": "filter[submitted]",
}, func(objectName string) string {
	return objectName
},
)

var PaginationObject = datautils.NewSet( //nolint:gochecknoglobals
	"scorecards",
	"playlists",
)

const objectEngagement = "engagements"

// MarshalledData
// chorus object has a special field named "attributes" which holds all the important fields.
// Therefore, nested "values" will be removed and fields inside the "values" field will be moved
// to the top level of the object.
//
// Example companies(shortened response):
//
//			{
//	  "data": [
//	    {
//	      "attributes": {
//	        "language": "string",
//	        "name": "string",
//	        "users": [
//	          5197,
//	          9449
//	        ],
//	        "default": true,
//	        "description": "string",
//	        "include_descendant_teams": false,
//	        "manager": 5078,
//	        "parent_team": 1232
//	      },
//	      "type": "team",
//	      "id": "123"
//	    },
//	    ...........
//	  ]
//	}
//
// The resulting fields for the above will be: type, id, language, name...
type MarshalledData func([]map[string]any, []string) ([]common.ReadResultRow, error)

func DataMarshall(resp *common.JSONHTTPResponse, nodePath string) MarshalledData {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		node, ok := resp.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		arr, err := jsonquery.New(node).ArrayOptional(nodePath)
		if err != nil {
			return nil, err
		}

		// No need to flatten records for the engagements object because its data is not nested under any nodePath.
		if nodePath == objectEngagement {
			arrData, err := jsonquery.Convertor.ArrayToMap(arr)
			if err != nil {
				return nil, err
			}

			return common.GetMarshaledData(arrData, fields)
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
		id, ok := record["id"].(string) // nolint:varnamelen
		if !ok {
			return nil, common.ErrEmptyRecordIdResponse
		}

		fieldRecord, ok := flattenRecords[id].(map[string]any)
		if !ok {
			return nil, common.ErrEmptyRecordIdResponse
		}

		data[i].Raw = record
		data[i].Fields = common.ExtractLowercaseFieldsFromRaw(fields, fieldRecord)
	}

	return data, nil
}

func flattenRecord(arr *ajson.Node) (map[string]any, error) {
	const keyValuesObject = "attributes"

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

		recordId, err := jsonquery.New(element).StringRequired("id")
		if err != nil {
			return nil, err
		}

		flattenMap[recordId] = flattenedRecord
	}

	return flattenMap, nil
}

func makeNextRecord(nextPage int, nodePath string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired(nodePath)
		if err != nil {
			return "", err
		}

		if len(value) == 0 {
			return "", nil
		}

		if nodePath == objectEngagement {
			continuationKey, err := jsonquery.New(node).StringOptional("continuation_key")
			if err != nil {
				return "", err
			}

			return *continuationKey, nil
		}

		nextStart := nextPage + PageSize

		return strconv.Itoa(nextStart), nil
	}
}

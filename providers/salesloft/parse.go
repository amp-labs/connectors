package salesloft

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
Response example:

	{
	  "metadata": {
	    "filtering": {},
	    "paging": {
	      "per_page": 25,
	      "current_page": 1,
	      "next_page": 2,
	      "prev_page": null
	    },
	    "sorting": {
	      "sort_by": "updated_at",
	      "sort_direction": "DESC NULLS LAST"
	    }
	  },
	  "data": [...]
	}
*/

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).ArrayRequired("data")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPageNum, err := jsonquery.New(node, "metadata", "paging").IntegerOptional("next_page")
		if err != nil {
			if errors.Is(err, jsonquery.ErrKeyNotFound) {
				// list resource doesn't support pagination, hence no next page
				return "", nil
			}

			return "", err
		}

		if nextPageNum == nil {
			// next page doesn't exist
			return "", nil
		}

		// use request URL to infer the next page URL
		reqLink.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))

		return reqLink.String(), nil
	}
}

// GetMarshalledDataWithIntId is very similar to common.GetMarshalledDataWithId, but handles the case where
// the "id" field is a numeric integer (float64) rather than a string, converting it to a string for the result.
func GetMarshalledDataWithIntId(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	fields = append(fields, "id")

	//nolint:varnamelen
	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}

		idAny := data[i].Fields["id"]
		if idAny == nil {
			return nil, errMissingId
		}

		intId, ok := idAny.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: %T", errUnexpectedIdType, idAny)
		}

		data[i].Id = strconv.FormatInt(int64(intId), 10)
	}

	return data, nil
}

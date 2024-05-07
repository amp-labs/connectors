package salesloft

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/linkutils"
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
func getTotalSize(node *ajson.Node) (int64, error) {
	return common.JSONManager.ArrSize(node, "data")
}

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := common.JSONManager.GetArr(node, "data")
	if err != nil {
		return nil, err
	}

	return common.JSONManager.ArrToMap(arr)
}

func makeNextRecordsURL(reqLink *linkutils.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nested, err := common.JSONManager.GetNestedObject(node, "metadata", "paging")
		if err != nil {
			if errors.Is(err, common.ErrKeyNotFound) {
				// list resource doesn't support pagination, hence no next page
				return "", nil
			}

			return "", err
		}

		nextPageNum, err := common.JSONManager.GetInteger(nested, "next_page", true)
		if err != nil {
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

func getMarshaledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

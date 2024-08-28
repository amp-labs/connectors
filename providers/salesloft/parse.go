package salesloft

import (
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
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
	arr, err := jsonquery.New(node).Array("data", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPageNum, err := jsonquery.New(node, "metadata", "paging").Integer("next_page", true)
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

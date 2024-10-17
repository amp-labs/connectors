package pipedrive

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Example Response Data
//
// {
//	"success":true,
//  "data":[{"id":8,"company_id":22122,"user_id":1234,"done":false,"type":"deadline"...}],
//  "additional_data":
// 		{
//		   "pagination":{
//	     	  "start":0,
// 		   	  "limit":100,
// 		      "more_items_in_collection":false,
//            "next_start":1
//    	}
// 	 }
// }
//

// nextRecordsURL builds the next-page url func.
func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// check if there is more items in the collection.
		more, err := jsonquery.New(node, "additional_data", "pagination").Bool("more_items_in_collection", true)
		if err != nil {
			return "", err
		}

		startValue, err := jsonquery.New(node, "additional_data", "pagination").Integer("next_start", true)
		if err != nil {
			return "", err
		}

		if *more {
			url.WithQueryParam("start", strconv.FormatInt(*startValue, 10))

			return url.String(), nil
		}

		return "", nil
	}
}

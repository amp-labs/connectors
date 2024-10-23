package zohocrm

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

/*
   "data": {...},
   "info": {
        "call": false,
        "per_page": 5,
        "next_page_token": "c8582xx9e7c7",
        "count": 5,
        "sort_by": "id",
        "page": 1,
        "previous_page_token": null,
        "page_token_expiry": "2022-11-11T15:08:14+05:30",
        "sort_order": "desc",
        "email": false,
        "more_records": true
    }
*/

func getNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		more, err := jsonquery.New(node, "info").Bool("more_records", false)
		if err != nil {
			return "", err
		}

		if *more {
			nextPageToken, err := jsonquery.New(node, "info").Str("next_page_token", true)
			if err != nil {
				return "", err
			}

			currPage, err := jsonquery.New(node, "info").Integer("page", false)
			if err != nil {
				return "", err
			}

			nextPage := *currPage + 1

			url.WithQueryParam("page_token", *nextPageToken)
			url.WithQueryParam("page", strconv.FormatInt(nextPage, 10))

			return url.String(), nil
		}

		return "", nil
	}
}

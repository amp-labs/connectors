package zoho

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
	doc: https://www.zoho.com/crm/developer/docs/api/v6/get-records.html
	The info object is not necessary required in zoho Desk
	Response Sample:

   "data": [{...}, {...}],
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
		hasMoreRecords, err := jsonquery.New(node, "info").BoolOptional("more_records")
		if err != nil {
			return "", err
		}

		if hasMoreRecords != nil && *hasMoreRecords {
			pageToken, err := jsonquery.New(node, "info").StringOptional("next_page_token")
			if err != nil {
				return "", err
			}

			url.WithQueryParam("page_token", *pageToken)

			return url.String(), nil
		}

		return "", nil
	}
}

func extractRecordsFromPath(objectName string) common.RecordsFunc {
	if objectName == users {
		return common.ExtractRecordsFromPath(users)
	}

	return common.ExtractRecordsFromPath("data")
}

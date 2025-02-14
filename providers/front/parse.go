package front

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
Sample Status OK Response
{
	"_links": {...},
	"_pagination":{
		"next":null,
	},
	"_results":{...}
}
*/

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonQuery := jsonquery.New(node, paginationResultKey)

		nextURL, err := jsonQuery.Str(nextURLKey, true)
		if err != nil {
			return "", err
		}

		// If  received null value,set the url to empty string
		if nextURL == nil {
			var emptyString string
			nextURL = &emptyString
		}

		return *nextURL, nil
	}
}

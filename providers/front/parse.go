package front

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey         = "limit"
	paginationResultKey = "_pagination"
	nextURLKey          = "next"
	pageSize            = "100"
	dataResponseKey     = "_results"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonQuery := jsonquery.New(node, paginationResultKey)

		nextURL, err := jsonQuery.StringOptional(nextURLKey)
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

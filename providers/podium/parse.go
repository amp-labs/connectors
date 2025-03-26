package podium

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		metadata, err := jsonquery.New(node).ObjectOptional("metadata")
		if err != nil {
			return "", err
		}

		nextPage, err := jsonquery.New(metadata).StringOptional("nextCursor")
		if err != nil {
			return "", err
		}

		if nextPage == nil {
			return "", nil
		}

		data, err := jsonquery.New(node).ArrayOptional("data")
		if err != nil {
			return "", err
		}

		// With some resources, the response would have the nextCursor value even in cases
		// where the next records list is empty. With this we check if the size of records
		// is equal to the size limit.Else there is no next-page.
		if len(data) < pageSize {
			return "", nil
		}

		return *nextPage, nil
	}
}

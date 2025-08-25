package xero

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectOptional("pagination")
		if err != nil || pagination == nil {
			return "", err
		}

		page, err := jsonquery.New(pagination).IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		pageCount, err := jsonquery.New(pagination).IntegerRequired("pageCount")
		if err != nil {
			return "", err
		}

		if page >= pageCount {
			return "", nil
		}

		nextPage := strconv.FormatInt(page+1, 10)

		return nextPage, nil
	}
}

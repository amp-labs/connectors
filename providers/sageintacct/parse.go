package sageintacct

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta, err := jsonquery.New(node).ObjectOptional("ia::meta")
		if err != nil || meta == nil {
			return "", nil //nolint: nilerr
		}

		totalCount, err := jsonquery.New(meta).IntegerRequired("totalCount")
		if err != nil {
			return "", nil //nolint: nilerr
		}

		nextStartPosition, err := jsonquery.New(meta).IntegerOptional("next")
		if err != nil || nextStartPosition == nil {
			return "", nil //nolint: nilerr
		}

		if totalCount < defaultPageSize {
			return "", nil
		}

		return strconv.Itoa(int(*nextStartPosition)), nil
	}
}

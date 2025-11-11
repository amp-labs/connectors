package happyfox

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta := jsonquery.New(node, "meta")

		currentPage, err := meta.IntegerOptional("page")
		if err != nil {
			return "", err
		}

		total, err := meta.IntegerOptional("totalPages")
		if err != nil {
			return "", err
		}

		if currentPage != nil || total != nil {
			if *currentPage < *total {
				return strconv.Itoa(int(*currentPage + 1)), nil
			}
		}

		return "", nil
	}
}

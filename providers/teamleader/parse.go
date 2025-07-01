package teamleader

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

func nextRecordsURL(params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		res, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		if res == nil {
			return "", nil
		}

		if len(res) < pageSize {
			return "", nil
		}

		currentPage := 1

		if params.NextPage != "" {
			page, err := strconv.Atoi(string(params.NextPage))
			if err == nil {
				currentPage = page
			}
		}

		return strconv.Itoa(currentPage + 1), nil
	}
}

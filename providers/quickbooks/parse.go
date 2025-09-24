package quickbooks

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		rcds, err := jsonquery.New(node, "QueryResponse").ArrayOptional(naming.CapitalizeFirstLetter(objectName))
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		startPosition, err := jsonquery.New(node, "QueryResponse").IntegerRequired("startPosition")
		if err != nil {
			return "", err
		}

		currentPageSize, err := jsonquery.New(node, "QueryResponse").IntegerRequired("maxResults")
		if err != nil {
			return "", err
		}

		pageSizeInt64, err := strconv.ParseInt(pageSize, 10, 64)
		if err != nil {
			return "", err
		}

		if currentPageSize < pageSizeInt64 {
			return "", nil
		}

		nextStartPosition := startPosition + pageSizeInt64

		return strconv.Itoa(int(nextStartPosition)), nil
	}
}

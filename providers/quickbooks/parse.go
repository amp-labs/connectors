package quickbooks

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getNodeRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		queryResponse, err := jsonquery.New(node).ObjectRequired("QueryResponse")

		if err != nil || queryResponse == nil {
			return nil, err
		}

		responseKey := objectNameToResponseField.Get(objectName)

		rcds, err := jsonquery.New(queryResponse).ArrayOptional(naming.CapitalizeFirstLetter(responseKey))
		if err != nil {
			return nil, err
		}

		if rcds == nil {
			return []*ajson.Node{}, nil
		}

		return rcds, nil
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		queryRes, err := jsonquery.New(node).ObjectRequired("QueryResponse")
		if err != nil || queryRes == nil {
			return "", err
		}

		startPosition, err := jsonquery.New(queryRes).IntegerOptional("startPosition")
		if err != nil || startPosition == nil {
			return "", err
		}

		currentPageSize, err := jsonquery.New(queryRes).IntegerRequired("maxResults")
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

		nextStartPosition := *startPosition + pageSizeInt64

		return strconv.Itoa(int(nextStartPosition)), nil
	}
}

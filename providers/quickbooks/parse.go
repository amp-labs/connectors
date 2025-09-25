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
		queryResponse, err := jsonquery.New(node).ObjectRequired("QueryResponse")

		if err != nil || queryResponse == nil {
			return nil, err
		}

		if objectName == "creditCardPayment" {
			// response key of creditCardPayment is CreditCardPaymentTxn
			objectName = "CreditCardPaymentTxn"
		}

		rcds, err := jsonquery.New(queryResponse).ArrayRequired(naming.CapitalizeFirstLetter(objectName))
		if err != nil || rcds == nil {
			// some endpoints return an empty object instead of an empty array when there are no records
			return []map[string]any{}, nil //nolint:nilerr
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
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

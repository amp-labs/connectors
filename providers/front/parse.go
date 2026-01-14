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

		return jsonQuery.StrWithDefault(nextURLKey, "")
	}
}

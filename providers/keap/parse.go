package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(moduleID common.ModuleID) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if moduleID == ModuleV1 {
			return jsonquery.New(node).StrWithDefault("next", "")
		}

		return jsonquery.New(node).StrWithDefault("next_page_token", "")
	}
}

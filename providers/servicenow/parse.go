package servicenow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func getNextRecordsURL(linkHeader string) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return common.ParseNexPageLinkHeader(linkHeader), nil
	}
}

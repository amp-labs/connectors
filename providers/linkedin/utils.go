package linkedin

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

var ObjectsWithSearchQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
)

var ObjectWithAccountId = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
)

func makeNextRecord() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return "", nil
	}
}

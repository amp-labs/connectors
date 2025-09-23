package linkedin

import "github.com/amp-labs/connectors/internal/datautils"

var ObjectsWithSearchQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
)

var ObjectWithAccountId = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
)

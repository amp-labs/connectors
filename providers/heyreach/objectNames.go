package heyreach

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameCampaign  = "campaign"
	objectNameLiAccount = "li_account"
	objectNameList      = "list"
)

var supportedObjectsByMetadata = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
	objectNameLiAccount,
	objectNameList,
)

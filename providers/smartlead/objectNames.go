package smartlead

import "github.com/amp-labs/connectors/common/handy"

const (
	objectNameCampaign     = "campaigns"
	objectNameEmailAccount = "email-accounts"
	objectNameClient       = "client"
)

var supportedObjectsByWrite = handy.NewSet([]string{ //nolint:gochecknoglobals
	objectNameCampaign,
	objectNameEmailAccount,
	objectNameClient,
})

var supportedObjectsByDelete = handy.NewSet([]string{ //nolint:gochecknoglobals
	objectNameCampaign,
})

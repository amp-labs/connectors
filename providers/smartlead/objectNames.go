package smartlead

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/smartlead/metadata"
)

const (
	objectNameCampaign     = "campaigns"
	objectNameEmailAccount = "email-accounts"
	objectNameClient       = "client"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = handy.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
	objectNameEmailAccount,
	objectNameClient,
)

var supportedObjectsByDelete = handy.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
)

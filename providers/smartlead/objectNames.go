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
var supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

var supportedObjectsByWrite = handy.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
	objectNameEmailAccount,
	objectNameClient,
)

var supportedObjectsByDelete = handy.NewSet( //nolint:gochecknoglobals
	// Removing campaign is the only to be supported at this time.
	// https://api.smartlead.ai/reference/delete-campaign
	objectNameCampaign,
)

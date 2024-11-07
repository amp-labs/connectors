package smartlead

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/smartlead/metadata"
)

const (
	objectNameCampaign     = "campaigns"
	objectNameEmailAccount = "email-accounts"
	objectNameClient       = "client"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
	objectNameEmailAccount,
	objectNameClient,
)

var supportedObjectsByDelete = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCampaign,
)

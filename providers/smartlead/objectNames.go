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

var (
	// Supported object names can be found under schemas.json.
	supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
		metadata.Schemas.GetObjectNames(),
	)
	supportedObjectsByWrite  = handy.MergeSets(createObjects.KeySet(), updateObjects.KeySet()) //nolint:gochecknoglobals
	supportedObjectsByDelete = deleteObjects.KeySet()                                          //nolint:gochecknoglobals
)

var createObjects = handy.Map[string, string]{ //nolint:gochecknoglobals
	// Create campaign.
	// https://api.smartlead.ai/reference/create-campaign
	objectNameCampaign: objectNameCampaign + "/create",
	// Create account.
	// https://api.smartlead.ai/reference/create-an-email-account
	objectNameEmailAccount: objectNameCampaign + "/save",
	// Add new client to the system.
	// https://api.smartlead.ai/reference/add-client-to-system-whitelabel-or-not
	objectNameClient: objectNameClient + "/save",
}

var updateObjects = handy.Map[string, string]{ // nolint:gochecknoglobals
	// Update account.
	// https://api.smartlead.ai/reference/update-email-account
	// It uses POST with RecordID.
	objectNameEmailAccount: objectNameEmailAccount,
}

var deleteObjects = handy.Map[string, string]{ //nolint:gochecknoglobals
	// Removing campaign is the only to be supported at this time.
	// https://api.smartlead.ai/reference/delete-campaign
	objectNameCampaign: objectNameCampaign,
}

var writeResponseRecordIdPaths = map[string]string{ // nolint:gochecknoglobals
	objectNameCampaign:     "id",
	objectNameEmailAccount: "emailAccountId",
	objectNameClient:       "clientId",
}

package gotocore

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectConfig describes how to fetch a sample record for a GoTo object on
// the admin (api.getgo.com) host. Both metadata and (eventually) read
// operations consult this registry.
type objectConfig struct {
	// path is the URL template under the BaseURL. The literal {accountKey}
	// is substituted with the connector's account key at resolve time.
	path string

	// recordsKey is the key inside the _embedded wrapper that holds the
	// records array. Empty falls back to the object name.
	recordsKey string
}

const accountKeyPlaceholder = "{accountKey}"

// objectRegistry maps object names to their endpoint metadata.
// GoTo Webinar v2 paths live under https://api.getgo.com/G2W/rest/v2/.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals
	"webinars":           {path: "G2W/rest/v2/organizers/{accountKey}/webinars"},
	"sessions":           {path: "G2W/rest/v2/organizers/{accountKey}/sessions"},
	"historicalMeetings": {path: "G2M/rest/historicalMeetings"},
	"upcomingMeetings":   {path: "G2M/rest/upcomingMeetings"},

	// For webhooks and userSubscriptions, the productType query parameter is required and must be set to "g2w" to retrieve webinar webhooks.
	// Ref: https://developer.goto.com/GoToWebinarV2#tag/Webhooks/operation/getWebhooks
	"webhooks":          {path: "G2W/rest/v2/webhooks?productType=g2w"},
	"userSubscriptions": {path: "G2W/rest/v2/userSubscriptions?productType=g2w"},
	"representatives":   {path: "G2AC/rest/v1/representatives"},
	"teams":             {path: "G2AC/rest/v1/teams/pages"},
	"portals":           {path: "G2AC/rest/v1/portals/pages"},
}

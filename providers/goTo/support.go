package goTo

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectConfig describes how to fetch a sample record for a GoTo Webinar object.
// Both metadata and (eventually) read operations consult this registry.
type objectConfig struct {
	// path is the URL template under the BaseURL. The literal {organizerKey} is
	// substituted with the connector's organizer key at resolve time.
	path string

	// recordsKey is the key inside the _embedded wrapper that holds the records array.
	// Empty falls back to the object name.
	recordsKey string
}

const organizerKeyPlaceholder = "{organizerKey}"

// objectRegistry maps object names to their endpoint metadata.
// GoTo Webinar v2 paths live under https://api.getgo.com/G2W/rest/v2/.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals
	"webinars": {path: "G2W/rest/v2/organizers/{organizerKey}/webinars"},
	"sessions": {path: "G2W/rest/v2/organizers/{organizerKey}/sessions"},
}

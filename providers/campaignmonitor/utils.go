package campaignmonitor

import (
	"fmt"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

var DirectEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"clients",
	"admins",
)

var EndpointsWithClientId = datautils.NewSet( //nolin:gochecknoglobals
	"lists",
	"segments",
	"suppressionlist",
	"templates",
	"people",
	"tags",
	"campaigns",
	"scheduled",
	"drafts",
	"journeys",
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	// Endpoint with clinet id in the url
	if EndpointsWithClientId.Has(objName) {
		objName = fmt.Sprintf("clients/%s/%s.json", c.clientID, objName)
	}

	// Endpoint without client id in the url.
	if DirectEndpoints.Has(objName) {
		objName += ".json"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, objName)
	if err != nil {
		return nil, err
	}

	return url, nil
}

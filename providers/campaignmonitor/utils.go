package campaignmonitor

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 1000

var DirectEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"clients",
	"admins",
)

var endpointsWithClientId = datautils.NewSet( //nolint:gochecknoglobals
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

var endpointsWtihResultsPath = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns",
	"suppressionlist",
)

var endpointsWithClientIdAfterObjName = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns",
	"templates",
	"lists",
)

var writeEndpointsWithClientId = datautils.NewSet( //nolint:gochecknoglobals
	"suppress",
	"credits",
	"sendingdomains",
	"people",
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	// Endpoint with client id in the url
	if endpointsWithClientId.Has(objName) {
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

func (c *Connector) constructWriteURL(objName string) (*urlbuilder.URL, error) {
	switch {
	case endpointsWithClientIdAfterObjName.Has(objName):
		objName = fmt.Sprintf("%s/%s.json", objName, c.clientID)

	case writeEndpointsWithClientId.Has(objName):
		objName = fmt.Sprintf("clients/%s/%s.json", c.clientID, objName)

	case DirectEndpoints.Has(objName):
		objName += ".json"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", APIVersion, objName)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func makeNextRecordsURL(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		url, err := urlbuilder.FromRawURL(reqLink)
		if err != nil {
			return "", err
		}

		numberOfPages, err := jsonquery.New(node).IntegerWithDefault("NumberOfPages", 0)
		if err != nil {
			return "", err
		}

		if numberOfPages != 0 {
			pageNumber, err := jsonquery.New(node).IntegerRequired("PageNumber")
			if err != nil {
				return "", err
			}

			if pageNumber != numberOfPages {
				pageNumber += 1

				url.WithQueryParam("page", strconv.Itoa(int(pageNumber)))

				return url.String(), nil
			}
		}

		return "", nil
	}
}

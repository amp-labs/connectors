package instantlyai

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const DefaultPageSize = 100

// directResponseEndpoints variable contains endpoint having direct
// response which means response not included any objects like "items".
var directResponseEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns/analytics",
	"campaigns/analytics/daily",
	"campaigns/analytics/steps",
)

var postEndpointsOfRead = datautils.NewSet( //nolint:gochecknoglobals
	"leads/list",
)

// https://developer.instantly.ai/api/v2/analytics/getdailycampaignanalytics
var sinceSupportedEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns/analytics/daily",
	"campaigns/analytics/steps",
)

// https://developer.instantly.ai/api/v2/analytics/getdailycampaignanalytics
var untilSupportedEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns/analytics/daily",
	"campaigns/analytics/steps",
)

func makeNextRecordsURL(reqLink *url.URL, objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if directResponseEndpoints.Has(objName) {
			return "", nil
		}

		url, err := urlbuilder.FromRawURL(reqLink)
		if err != nil {
			return "", err
		}

		pagination, err := jsonquery.New(node).StringRequired("next_starting_after")
		if err != nil {
			return "", err
		}

		if pagination != "" {
			url.WithQueryParam("starting_after", pagination)

			return url.String(), nil
		}

		return "", nil
	}
}

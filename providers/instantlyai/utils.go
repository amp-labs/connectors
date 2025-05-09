package instantlyai

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const DefaultPageSize = 100

// directResponseEndpoints variable contains endpoint having direct
// response which means response not included any objects like "items".
var directResponseEndpoints = datautils.NewSet( //nolint:gochecknoglobals
	"campaigns/analytics",
	"campaigns/analytics/overview",
	"campaigns/analytics/daily",
	"campaigns/analytics/steps",
)

var postEndpointsOfRead = datautils.NewSet( //nolint:gochecknoglobals
	"leads/list",
)

func makeNextRecordsURL(reqLink *url.URL, objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if directResponseEndpoints.Has(objName) {
			return "", nil
		}

		pagination, err := jsonquery.New(node).StringRequired("next_starting_after")
		if err != nil {
			return "", err
		}

		if pagination != "" {
			nextLink := *reqLink
			query := nextLink.Query()
			query.Set("starting_after", pagination)
			nextLink.RawQuery = query.Encode()

			return nextLink.String(), nil
		}

		return "", nil
	}
}

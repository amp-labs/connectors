package linkedin

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const pageSize = 100

var ObjectsWithSearchQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
)

var ObjectWithAccountId = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
)

func makeNextRecord(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		url, err := urlbuilder.FromRawURL(reqLink)
		if err != nil {
			return "", err
		}

		pagination, err := jsonquery.New(node).ObjectOptional("metadata")
		if err != nil {
			return "", err
		}

		if pagination != nil {
			nextPage, err := jsonquery.New(pagination).StrWithDefault("nextPageToken", "")
			if err != nil {
				return "", err
			}

			if nextPage == "" {
				return "", nil
			}

			url.WithQueryParam("pageToken", nextPage)

			return url.String(), nil
		}

		return "", nil
	}
}

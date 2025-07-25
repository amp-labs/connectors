package lever

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	defaultPageSize = 100
)

var (
	endpointWithCreatedAtRange = datautils.NewSet( //nolint:gochecknoglobals
		"audit_events",
		"requisitions",
	)

	endpointWithUpdatedAtRange = datautils.NewSet( //nolint:gochecknoglobals
		"postings",
		"opportunities",
	)
)

func makeNextRecordsURL(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		url, err := urlbuilder.FromRawURL(reqLink)
		if err != nil {
			return "", err
		}

		hasNextPage, err := jsonquery.New(node).BoolWithDefault("hasNext", false)
		if err != nil {
			return "", err
		}

		if hasNextPage {
			pagination, err := jsonquery.New(node).StringRequired("next")
			if err != nil {
				return "", err
			}

			url.WithQueryParam("offset", pagination)

			return url.String(), nil
		}

		return "", nil
	}
}

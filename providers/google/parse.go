package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		pageToken, err := jsonquery.New(node).StrWithDefault("nextPageToken", "")
		if err != nil {
			return "", err
		}

		if len(pageToken) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		url.AddEncodingExceptions(map[string]string{
			"%3D": "=",
		})
		url.WithQueryParam("pageToken", pageToken)

		return url.String(), nil
	}
}

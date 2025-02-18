package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func getNextRecordURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageToken, err := jsonquery.New(node).StrWithDefault("next_page_token", "")
		if err != nil {
			return "", err
		}

		if len(pageToken) == 0 {
			return "", nil
		}

		reqLink.WithQueryParam("next_page_token", pageToken)

		return reqLink.String(), nil
	}
}

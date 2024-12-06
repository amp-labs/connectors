package constantcontact

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		href, err := jsonquery.New(node, "_links", "next").StrWithDefault("href", "")
		if err != nil {
			return "", err
		}

		if len(href) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		fullURL := baseURL + href

		return fullURL, nil
	}
}

package constantcontact

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

		url, err := urlbuilder.New(fullURL)
		if err != nil {
			return "", err
		}

		// Cursor is base64 encoded,
		// therefore it may contain symbols that should be exempt from escaping.
		url.AddEncodingExceptions(map[string]string{
			"%3D": "=",
		})

		return url.String(), nil
	}
}

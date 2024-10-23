package customerapp

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		next, err := jsonquery.New(node).StrWithDefault("next", "")
		if err != nil {
			return "", err
		}

		if len(next) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		// Next page token is base64 encoded,
		// therefore it may contain symbols that should be exempt from escaping.
		reqLink.AddEncodingExceptions(map[string]string{
			"%3D": "=",
		})

		reqLink.WithQueryParam("start", next)
		reqLink.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

		return reqLink.String(), nil
	}
}
